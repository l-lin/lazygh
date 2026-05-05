package tui

const treeSitterJavaScriptHighlightsQuery = `; Functions

(function_expression
  name: (identifier) @function)
(function_declaration
  name: (identifier) @function)
(method_definition
  name: (property_identifier) @function.method)

(call_expression
  function: (identifier) @function)
(call_expression
  function: (member_expression
    property: (property_identifier) @function.method))

; Properties

(property_identifier) @property

; Special identifiers

((identifier) @constructor
 (#match? @constructor "^[A-Z]"))

; Literals

(comment) @comment

[
  (string)
  (template_string)
] @string

(number) @number

[
  (true)
  (false)
  (null)
  (undefined)
] @constant.builtin

; Keywords

[
  "as"
  "async"
  "await"
  "break"
  "case"
  "catch"
  "class"
  "const"
  "continue"
  "debugger"
  "default"
  "delete"
  "do"
  "else"
  "export"
  "extends"
  "finally"
  "for"
  "from"
  "function"
  "get"
  "if"
  "import"
  "in"
  "instanceof"
  "let"
  "new"
  "of"
  "return"
  "set"
  "static"
  "switch"
  "target"
  "throw"
  "try"
  "typeof"
  "var"
  "void"
  "while"
  "with"
  "yield"
] @keyword
`

const treeSitterTypeScriptHighlightsSupplementQuery = `; Types

(type_identifier) @type
(predefined_type) @type.builtin

((identifier) @type
 (#match? @type "^[A-Z]"))

; Keywords

[
  "abstract"
  "declare"
  "enum"
  "export"
  "implements"
  "interface"
  "keyof"
  "namespace"
  "private"
  "protected"
  "public"
  "readonly"
  "override"
  "satisfies"
  "type"
] @keyword
`

const treeSitterTypeScriptHighlightsQuery = treeSitterJavaScriptHighlightsQuery + "\n" + treeSitterTypeScriptHighlightsSupplementQuery

const treeSitterHTMLHighlightsQuery = `(tag_name) @tag
(erroneous_end_tag_name) @tag.error
(doctype) @constant
(attribute_name) @attribute
(attribute_value) @string
(comment) @comment
`

const treeSitterXMLHighlightsQuery = `; XML declaration

"xml" @keyword

[ "version" "encoding" "standalone" ] @property

(EncName) @string.special
(VersionNum) @number

[ "yes" "no" ] @boolean

; Tags

(STag (Name) @tag)
(ETag (Name) @tag)
(EmptyElemTag (Name) @tag)

; Attributes

(Attribute (Name) @property)
(Attribute (AttValue) @string)

; Entities and text

(EntityRef) @constant
((EntityRef) @constant.builtin
 (#any-of? @constant.builtin
   "&amp;" "&lt;" "&gt;" "&quot;" "&apos;"))

(CharData) @markup

; Misc

(Comment) @comment
`

const treeSitterKotlinHighlightsQuery = `; Functions

(function_declaration
  name: (identifier) @function)

(call_expression
  (identifier) @function)

; Types

(user_type
  (identifier) @type)

; Properties

(variable_declaration
  (identifier) @property)

; Literals

[
  (string_literal)
  (multiline_string_literal)
  (character_literal)
] @string

(number_literal) @number

[
  (block_comment)
  (line_comment)
] @comment

`

const treeSitterLuaHighlightsQuery = `; Keywords

"return" @keyword.return

[
  "goto"
  "in"
  "local"
  "global"
] @keyword

(break_statement) @keyword

(do_statement
  [
    "do"
    "end"
  ] @keyword)

(while_statement
  [
    "while"
    "do"
    "end"
  ] @repeat)

(repeat_statement
  [
    "repeat"
    "until"
  ] @repeat)

(if_statement
  [
    "if"
    "elseif"
    "else"
    "then"
    "end"
  ] @conditional)

(elseif_statement
  [
    "elseif"
    "then"
    "end"
  ] @conditional)

(else_statement
  [
    "else"
    "end"
  ] @conditional)

(for_statement
  [
    "for"
    "do"
    "end"
  ] @repeat)

(function_declaration
  [
    "function"
    "end"
  ] @keyword.function)

(function_definition
  [
    "function"
    "end"
  ] @keyword.function)

; Properties

(field
  name: (identifier) @field)

(dot_index_expression
  field: (identifier) @field)

; Functions

(function_declaration
  name: [
    (identifier) @function
    (dot_index_expression
      field: (identifier) @function)
  ])

(function_declaration
  name: (method_index_expression
    method: (identifier) @method))

(function_call
  name: [
    (identifier) @function.call
    (dot_index_expression
      field: (identifier) @function.call)
    (method_index_expression
      method: (identifier) @method.call)
  ])

(function_call
  (identifier) @function.builtin
  (#any-of? @function.builtin
    "assert" "collectgarbage" "dofile" "error" "getfenv" "getmetatable" "ipairs" "load" "loadfile"
    "loadstring" "module" "next" "pairs" "pcall" "print" "rawequal" "rawget" "rawset" "require"
    "select" "setfenv" "setmetatable" "tonumber" "tostring" "type" "unpack" "xpcall"))

; Literals

(comment) @comment
(hash_bang_line) @preproc
(number) @number
(string) @string
(escape_sequence) @string.escape
`

const treeSitterMakeHighlightsQuery = `[
  (text)
  (string)
  (raw_text)
] @string

(variable_assignment
  name: (word) @constant)

(variable_reference
  (word) @constant)

(comment) @comment

[
  "ifeq"
  "ifneq"
  "ifdef"
  "ifndef"
  "else"
  "endif"
  "if"
  "or"
  "and"
] @conditional

"foreach" @repeat

[
  "define"
  "endef"
  "vpath"
  "undefine"
  "export"
  "unexport"
  "override"
  "private"
] @keyword

[
  "include"
  "sinclude"
  "-include"
] @include

[
  "subst"
  "patsubst"
  "strip"
  "findstring"
  "filter"
  "filter-out"
  "sort"
  "word"
  "words"
  "wordlist"
  "firstword"
  "lastword"
  "dir"
  "notdir"
  "suffix"
  "basename"
  "addsuffix"
  "addprefix"
  "join"
  "wildcard"
  "realpath"
  "abspath"
  "call"
  "eval"
  "file"
  "value"
  "shell"
] @keyword.function

[
  "error"
  "warning"
  "info"
] @exception
`

const treeSitterPythonHighlightsQuery = `; Identifier naming conventions

((identifier) @constructor
 (#match? @constructor "^[A-Z]"))

; Function calls

(decorator) @function
(decorator
  (identifier) @function)

(call
  function: (attribute attribute: (identifier) @function.method))
(call
  function: (identifier) @function)

; Function definitions

(function_definition
  name: (identifier) @function)

(attribute attribute: (identifier) @property)
(type (identifier) @type)

; Literals

[
  (none)
  (true)
  (false)
] @constant.builtin

[
  (integer)
  (float)
] @number

(comment) @comment
(string) @string
(escape_sequence) @escape

[
  "as"
  "assert"
  "async"
  "await"
  "break"
  "class"
  "continue"
  "def"
  "del"
  "elif"
  "else"
  "except"
  "exec"
  "finally"
  "for"
  "from"
  "global"
  "if"
  "import"
  "lambda"
  "nonlocal"
  "pass"
  "print"
  "raise"
  "return"
  "try"
  "while"
  "with"
  "yield"
  "match"
  "case"
] @keyword
`

const treeSitterRubyHighlightsQuery = `(identifier) @variable

[
  "alias"
  "and"
  "begin"
  "break"
  "case"
  "class"
  "def"
  "do"
  "else"
  "elsif"
  "end"
  "ensure"
  "for"
  "if"
  "in"
  "module"
  "next"
  "or"
  "rescue"
  "retry"
  "return"
  "then"
  "unless"
  "until"
  "when"
  "while"
  "yield"
] @keyword

((identifier) @keyword
 (#match? @keyword "^(private|protected|public)$"))

(constant) @constructor

; Function calls

"defined?" @function.method.builtin

(call
  method: [(identifier) (constant)] @function.method)

((identifier) @function.method.builtin
 (#eq? @function.method.builtin "require"))

; Function definitions

(alias (identifier) @function.method)
(setter (identifier) @function.method)
(method name: [(identifier) (constant)] @function.method)
(singleton_method name: [(identifier) (constant)] @function.method)

; Properties

[
  (class_variable)
  (instance_variable)
] @property

; Literals

[
  (string)
  (bare_string)
  (subshell)
  (heredoc_body)
  (heredoc_beginning)
] @string

[
  (simple_symbol)
  (delimited_symbol)
  (hash_key_symbol)
  (bare_symbol)
] @string.special.symbol

(regex) @string.special.regex
(escape_sequence) @escape

[
  (integer)
  (float)
] @number

[
  (nil)
  (true)
  (false)
] @constant.builtin

(comment) @comment
`
