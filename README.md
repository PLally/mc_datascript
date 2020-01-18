# Minecraft Datapack scripting language
### TODO
- print location of errors
- better error handling
- allow some sort of SLEEP and LOOP function
- allow calling functions from different namespaces
- atomically  generate tick.json file
- rewrite compile.go
- move lex.go and compile.go into separate packages
- add ALIAS support outside strings and commands 

### Info
- language keywords are allways uppercase e.g. `FUNC` `ALIAS` `SET` and are always followed by an identifier (a-z and `_`)
- `{` and `}` open and close a block, everything in a block will eventually be translated into a list of minecraft commands
- text enclosed in \` is interpreted as a minecraft command e.g. \`say "Hello World"\`
- `#{name}`  will look up any defined  `ALIAS` with that name, and `#{name}` will be replaced with the  `ALIAS`
