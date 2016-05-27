package dot_test

import (
	"fmt"
	"log"
)

import (
	"github.com/timtadh/dot"
	lex "github.com/timtadh/lexmachine"
)

func Example() {
	s, err := dot.Lexer.Scanner([]byte(`digraph {
  rankdir=LR;
  a [label="a" shape=box];
  c [<label>=<<u>C</u>>];
  b [label="bb"];
  a -> c;
  c -> b;
  d -> c;
  b -> a;
  b -> e;
  e -> f;
}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Type    | Lexeme     | Position")
	fmt.Println("--------+------------+------------")
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		if err != nil {
			log.Fatal(err)
		}
		token := tok.(*lex.Token)
		fmt.Printf("%-7v | %-10v | %v:%v-%v:%v\n", dot.Tokens[token.Type], string(token.Lexeme), token.StartLine, token.StartColumn, token.EndLine, token.EndColumn)
	}
	// Output:
	// Type    | Lexeme     | Position
	// --------+------------+------------
	// DIGRAPH | digraph    | 1:1-1:7
	// {       | {          | 1:9-1:9
	// ID      | rankdir    | 2:3-2:9
	// =       | =          | 2:10-2:10
	// ID      | LR         | 2:11-2:12
	// ;       | ;          | 2:13-2:13
	// ID      | a          | 3:3-3:3
	// [       | [          | 3:5-3:5
	// ID      | label      | 3:6-3:10
	// =       | =          | 3:11-3:11
	// ID      | "a"        | 3:12-3:14
	// ID      | shape      | 3:16-3:20
	// =       | =          | 3:21-3:21
	// ID      | box        | 3:22-3:24
	// ]       | ]          | 3:25-3:25
	// ;       | ;          | 3:26-3:26
	// ID      | c          | 4:3-4:3
	// [       | [          | 4:5-4:5
	// ID      | <label>    | 4:6-4:12
	// =       | =          | 4:13-4:13
	// ID      | <<u>C</u>> | 4:14-4:23
	// ]       | ]          | 4:24-4:24
	// ;       | ;          | 4:25-4:25
	// ID      | b          | 5:3-5:3
	// [       | [          | 5:5-5:5
	// ID      | label      | 5:6-5:10
	// =       | =          | 5:11-5:11
	// ID      | "bb"       | 5:12-5:15
	// ]       | ]          | 5:16-5:16
	// ;       | ;          | 5:17-5:17
	// ID      | a          | 6:3-6:3
	// ->      | ->         | 6:5-6:6
	// ID      | c          | 6:8-6:8
	// ;       | ;          | 6:9-6:9
	// ID      | c          | 7:3-7:3
	// ->      | ->         | 7:5-7:6
	// ID      | b          | 7:8-7:8
	// ;       | ;          | 7:9-7:9
	// ID      | d          | 8:3-8:3
	// ->      | ->         | 8:5-8:6
	// ID      | c          | 8:8-8:8
	// ;       | ;          | 8:9-8:9
	// ID      | b          | 9:3-9:3
	// ->      | ->         | 9:5-9:6
	// ID      | a          | 9:8-9:8
	// ;       | ;          | 9:9-9:9
	// ID      | b          | 10:3-10:3
	// ->      | ->         | 10:5-10:6
	// ID      | e          | 10:8-10:8
	// ;       | ;          | 10:9-10:9
	// ID      | e          | 11:3-11:3
	// ->      | ->         | 11:5-11:6
	// ID      | f          | 11:8-11:8
	// ;       | ;          | 11:9-11:9
	// }       | }          | 12:1-12:1
}
