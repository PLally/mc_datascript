package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type Parser struct {
	tokens              []token
	currentItem         token
	pos                 int
	namespace           string // current file namespace
	alias               map[string]string
	functions           []mcfunction
	anonymousConstantId int
	mcmeta              MCMeta
}

type mcfunction struct {
	namespace string
	filepath  string
	commands  []string
}

func (p *Parser) next() token {
	item := p.tokens[p.pos]
	p.currentItem = item
	p.pos++
	return item
}

type ParserStateFunc func(p *Parser) ParserStateFunc

func (p *Parser) run() {
	// todo this file needs to be rewritten
	state := parseFile
	for state != nil {
		state = state(p)
	}
	p.mcmeta.Pack.PackFormat = 5

	for _, mcf := range p.functions {

		filepath := "out/data/" + path.Join(mcf.namespace, "functions", mcf.filepath) + ".mcfunction"
		os.MkdirAll(path.Dir(filepath), 0777)
		data := []byte(strings.Join(mcf.commands, "\n"))
		err := ioutil.WriteFile(filepath, data, 0777)
		if err != nil {
			fmt.Println(err)
		}
	}
	pack_meta, err := json.Marshal(p.mcmeta)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("out/pack.mcmeta", pack_meta, 0777)
}

func (p *Parser) expandAliases(s string) string {
	aliasStart := -1
	output := ""
	for i := 0; i < len(s); i++ {
		if aliasStart <= -1 {
			c := s[i]
			if s[i] == '#' && s[i+1] == '{' {
				i += 2
				aliasStart = i
			} else {
				output += string(c)
			}
		} else {
			if s[i] == '}' {
				name := s[aliasStart:i]
				output += p.alias[name]
				aliasStart = -1
			}
		}
	}
	return output
}

func parseFile(p *Parser) ParserStateFunc {

	for {
		if p.pos >= len(p.tokens) {
			return nil
		}

		switch item := p.next(); item.itemType {

		case DefItem:
			return parseTopLevelDef
		case ErrorItem:
			fmt.Println(item.value)
			os.Exit(1)
		case EOFItem:
			return nil
		default:
			fmt.Println("Unrecognized token", item.itemType)
			return nil
		}
	}
}

func parseTopLevelDef(p *Parser) ParserStateFunc {
	if p.namespace == "" && p.currentItem.value != "NAMESPACE" {
		panic("You must define a namespace")
	}

	switch p.currentItem.value {
	case "NAMESPACE":
		p.namespace = p.next().value
		return parseFile
	case "FUNC":
		return parseFunc
	case "ALIAS":
		return parseAlias
	}
	fmt.Printf("Unrecognized token %v\n", p.currentItem.itemType)
	return nil
}

func parseFunc(p *Parser) ParserStateFunc {
	mcf := mcfunction{
		namespace: p.namespace,
		filepath:  p.next().value,
	}
	if p.next().itemType != StartBlockItem {
		panic("No block start after FUNC")
	}

	lines := parseBlock(p)
	mcf.commands = lines
	p.functions = append(p.functions, mcf)
	return parseFile
}

// parses a block until it encounters "}" returning the lines it contains
func parseBlock(p *Parser) (lines []string) {
	for {
		switch item := p.next(); item.itemType {
		case EndBlockItem:
			return lines
		case CommandItem:
			lines = append(lines, p.expandAliases(item.value[1:len(item.value)-1]))
		case DefItem:
			lines = append(lines, parseInnerDef(p)...)
		}
	}
}

func parseInnerDef(p *Parser) []string {
	switch p.currentItem.value {
	case "SET":
		return parseSET(p)
	case "CALL":
		return parseCALL(p)
	case "IF":
		return parseIF(p)
	default:
		panic(fmt.Sprintf("Unrecognized operation %v", p.currentItem.value))
	}
}

func parseIF(p *Parser) (lines []string) {
	funcName := "if/f" + strconv.Itoa(p.anonymousConstantId)
	boardName := p.namespace + "_vars"
	p.anonymousConstantId++
	name := p.next()
	condition := p.next()
	value := p.next()

	if p.next().itemType != StartBlockItem {
		panic("No block start after IF")
	}

	mcf := mcfunction{
		namespace: p.namespace,
		filepath:  funcName,
	}
	mcf.commands = parseBlock(p)
	p.functions = append(p.functions, mcf)

	val := value.value
	if value.itemType == IntegerItem {
		val = "anon_const_" + strconv.Itoa(p.anonymousConstantId)
		p.anonymousConstantId++
		lines = append(lines, fmt.Sprintf("scoreboard players set %v %v %v", val, boardName, value.value))
	}

	mcf.commands = append(mcf.commands)
	p.functions = append(p.functions, mcf)
	return append(lines, fmt.Sprintf(
		"execute if score %[1]v %[2]v %[3]v %[4]v %[2]v run function %[5]v:%[6]v",
		name.value, boardName, condition.value, val, mcf.namespace, mcf.filepath,
	))

}

func parseCALL(p *Parser) []string {
	funcName := p.namespace + ":" + p.next().value
	return []string{fmt.Sprintf("function %v", funcName)}
}

func parseSET(p *Parser) (lines []string) {
	name := p.next()
	operator := p.next()
	value := p.next()
	boardName := p.namespace + "_vars"
	switch operator.value {
	case "=":
		return []string{fmt.Sprintf("scoreboard players set %v %v %v", name.value, boardName, value.value)}
		return lines
	case "-=": 
		fallthrough
	case "+=":
		fallthrough
	case "*=":
		fallthrough
	case "/=":
		fallthrough
	case "%=":
		val := value.value
		if value.itemType == IntegerItem {
			val = "anon_const_" + strconv.Itoa(p.anonymousConstantId)
			p.anonymousConstantId++
			lines = append(lines, fmt.Sprintf("scoreboard players set %v %v %v", val, boardName, value.value))
		}
		return append(lines, fmt.Sprintf("scoreboard players operation %v %v %v %v %v", name.value, boardName, operator.value, val, boardName))

	default:
		panic("Unrecognized operator")
	}
}
func parseAlias(p *Parser) ParserStateFunc {
	name := p.next().value
	p.next()
	value := p.next()
	if value.itemType == StringItem {
		p.alias[name] = value.value[1 : len(value.value)-1]
	} else {
		p.alias[name] = value.value
	}
	return parseFile
}
