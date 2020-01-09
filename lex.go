package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	AlphaUpper         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	AlphaLower         = "abcdefghijklmnopqrstuvwxyz_"
	AlphaAll           = AlphaLower + AlphaUpper
	NumberCharacters   = "1234567890"
	OperatorCharacters = "-+*%/"
	WhiteSpace         = "\t\n "
)

type itemType int

const (
	ErrorItem itemType = iota
	DefItem
	IdentItem
	CommandItem
	StartBlockItem
	EndBlockItem
	AssignmentItem
	StringItem
	IntegerItem
	CommentItem
	ConditionItem
	EOFItem
)

type token struct {
	itemType
	value string
}

type lexer struct {
	pos         int
	start       int
	src         string
	items       []token
	exitOnError bool
}

func (l *lexer) next() (c uint8) {
	c = l.src[l.pos]
	l.pos += 1
	return c
}
func (l *lexer) accept(valid string) bool {
	if strings.Contains(valid, string(l.next())) {
		return true
	}
	l.backup()
	return false
}
func (l *lexer) backup() {
	l.pos -= 1
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(item itemType) {
	val := l.src[l.start:l.pos]
	l.start = l.pos
	l.items = append(l.items, token{item, val})
}

func (l *lexer) errorf(f string, args ...interface{}) {
	item := token{
		itemType: ErrorItem,
		value:    fmt.Sprintf(f, args...),
	}
	if l.exitOnError {
		fmt.Println(item.value)
		os.Exit(1)
	}
	l.items = append(l.items, item)
}
func (l *lexer) run() {
	state := lexMain
	for state != nil {
		state = state(l)
	}
}

type StateFunc func(*lexer) StateFunc

func lexMain(l *lexer) StateFunc {
	if l.pos >= len(l.src) {
		l.ignore()
		l.emit(EOFItem)
		return nil
	}

	switch {
	case l.accept(";"):
		return lexComment
	case l.accept(WhiteSpace):
		l.ignore()
	case l.accept(AlphaUpper):
		return lexDef
	case l.accept("{"):
		l.emit(StartBlockItem)
	case l.accept("}"):
		l.emit(EndBlockItem)
	case l.accept("="):
		l.emit(AssignmentItem)
	case l.accept("\""):
		return lexString
	case l.accept("`"):
		return lexCommand
	case l.accept("><"):
		l.accept("=")
		l.emit(ConditionItem)
		return lexAfterAssignment
	case l.accept("-"):
		if l.accept("=") {
			l.emit(AssignmentItem)
			return lexAfterAssignment
		}
		if !l.accept(NumberCharacters) {
			break
		}
		fallthrough
	case l.accept(NumberCharacters):
		return lexNumber
	case l.accept(OperatorCharacters):
		if l.accept("=") {
			l.emit(AssignmentItem)
			return lexAfterAssignment
		}

	default:
		fmt.Println(l.src[l.pos])
		fmt.Println(l.pos)
		l.errorf("Unrecognized character %v", l.src[l.start])
		return nil
	}

	return lexMain

}

func lexComment(l *lexer) StateFunc {
	if l.accept("\n") {
		l.backup()
		l.emit(CommentItem)
		return lexMain
	}

	l.next()
	return lexComment
}
func lexAfterAssignment(l *lexer) StateFunc {
	switch {
	case l.accept(WhiteSpace):
		l.ignore()
	case l.accept(NumberCharacters):
		return lexNumber
	case l.accept(AlphaAll):
		return lexIdent
	}
	return lexAfterAssignment
}
func lexCommand(l *lexer) StateFunc {
	if l.pos >= len(l.src) {
		l.ignore()
		l.errorf("Unterminated command")
		l.emit(EOFItem)
		return nil
	}

	switch {
	case l.accept("`"):
		l.emit(CommandItem)
		return lexMain
	default:
		l.next()
	}
	return lexCommand
}

func lexDef(l *lexer) StateFunc {
	isDef := l.accept(AlphaUpper) || l.pos >= len(l.src)
	if !isDef {
		l.emit(DefItem)
		if l.next() == ' ' {
			l.ignore()
		}
		return lexIdent
	}
	return lexDef
}

func lexIdent(l *lexer) StateFunc {
	if !l.accept(AlphaAll) || l.pos >= len(l.src) {
		l.emit(IdentItem)
		return lexMain
	}
	return lexIdent
}

func lexString(l *lexer) StateFunc {
	if l.pos >= len(l.src) {
		l.ignore()
		l.errorf("Unterminated String")
		l.emit(EOFItem)
		return nil
	}

	switch {
	case l.accept("\""):
		l.emit(StringItem)
		return lexMain
	default:
		l.next()
	}
	return lexString
}
func lexNumber(l *lexer) StateFunc {
	if l.pos >= len(l.src) {
		l.emit(IntegerItem)
		return nil
	}
	if l.accept(WhiteSpace) {
		l.backup()
		l.emit(IntegerItem)
		return lexMain
	}
	l.next()
	return lexNumber
}

func LexText(src string) []token {
	l := lexer{
		src:         src,
		exitOnError: true,
	}
	l.run()
	return l.items
}
