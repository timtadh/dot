# Dot

by Tim Henderson (tadh@case.edu)

Copyright 2016, Licensed under the GPL version 3. Please reach out to me
directly if you require another licensing option. I am willing to work with you.

## Purpose

A parser for the [graphviz dot
language](http://www.graphviz.org/doc/info/lang.html) in the Go programming
language. It is intended to be stream oriented for parsing large graphs.

## Grammar of Dot

Dot is a relatively simple language and can be parsed with a clean separation
between the lexing and parsing phases.

### The Token Types

```
NODE = "node"
EDGE = "edge"
GRAPH = "graph"
DIGRAPH = "digraph"
SUBGRAPH = "subgraph"
STRICT = "strict"
ID = ([a-zA-Z_][a-zA-Z0-9_]*)|ID-STRING|ID-HTML
LSQUARE = "["
RSQUARE = "]"
LCURLY = "{"
RCURLY = "}"
EQUAL = "="
COMMA = ","
SEMI = ";"
COLON = ":"
ARROW = "->"
DDASH = "--"
COMMENT = (/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/)|(//.*$)
```

Note: The `ID` token is has 3 forms:

1. The usual form as a name `[a-zA-Z_][a-zA-Z0-9_]*`

2. A string, which is actually has a non-regular specification:

        CHAR = [^\"]
        ID-STRING = IdString

        IdString : " Body " ;
        Body : Char Body ;
             | e                    // denotes epsilon, the empty string
             ;
        Char : CHAR
             | \ \
             | \ "
             | \ CHAR
             ;

    Thus `"\\\""` is valid but `"\\""` is not

2. A HTML string, which also is non-regular:

        CHAR = [^<>]
        ID-HTML = IdHTML

        IdHTML : Tag ;
        Tag : < Body > ;
        Body : CHAR Body ;
             | Tag Body
             | e                    // denotes epsilon, the empty string
             ;

    Thus `<<xyz<xy>xyz><asdf>>` is valid but `<<>` is not


### The Grammar

This is the grammar as given on <http://www.graphviz.org/doc/info/lang.html>

```
graph   :   [ strict ] (graph | digraph) [ ID ] '{' stmt_list '}'
stmt_list   :   [ stmt [ ';' ] [ stmt_list ] ]
stmt    :   node_stmt
    |   edge_stmt
    |   attr_stmt
    |   ID '=' ID
    |   subgraph
attr_stmt   :   (graph | node | edge) attr_list
attr_list   :   '[' [ a_list ] ']' [ attr_list ]
a_list  :   ID '=' ID [ (';' | ',') ] [ a_list ]
edge_stmt   :   (node_id | subgraph) edgeRHS [ attr_list ]
edgeRHS     :   edgeop (node_id | subgraph) [ edgeRHS ]
node_stmt   :   node_id [ attr_list ]
node_id     :   ID [ port ]
port        :   ':' ID [ ':' compass_pt ]
            |   ':' compass_pt
subgraph    :   [ subgraph [ ID ] ] '{' stmt_list '}'
compass_pt  :   (n | ne | e | se | s | sw | w | nw | c | _)
```

#### A LALR Grammar

A grammar suitable for bottom up parsing

```
Graphs : Graphs Graph
       | Graph
       ;

Graph : GraphStmt
      | COMMENT
      ;

GraphStmt : GraphType GraphBody
          | GraphType ID GraphBody
          | STRICT GraphType GraphBody
          | STRICT GraphType ID GraphBody
          ;

GraphType : GRAPH
          | DIGRAPH
          ;

GraphBody : LCURLY StmtList RCURLY
          | LCURLY RCURLY
          ;

StmtList : Stmt StmtList
         | Stmt SEMI StmtList
         | Stmt
         | Stmt SEMI
         ;

Stmt : ID EQUAL ID
     | NodeStmt
     | EdgeStmt
     | AttrStmt
     | SubGraph
     | COMMENT
     ;

NodeStmt : NodeId
         | NodeId AttrLists
         ;

NodeId : ID
       | ID Port
       ;

Port : COLON ID
     | COLON ID COLON ID // where second ID in "n", "ne", "e", "se", "s", "sw",
                         //                    "w", "nw", "c", "_"
     ;

AttrStmt : AttrType AttrLists
         ;

AttrType : NODE
         | EDGE
         | GRAPH
         ;

AttrLists : AttrList AttrLists
          | AttrList
          ;

AttrList : LSQUARE AttrExprs RSQUARE
         | LSQUARE RSQUARE
         ;

AttrExprs : AttrExpr AttrExprs
          | AttrExpr
          ;

AttrExpr : ID EQUAL ID
         | ID EQUAL ID COMMA
         | ID EQUAL ID SEMI
         ;

EdgeStmt : EdgeReciever EdgeRHS
         | EdgeReciever EdgeRHS AttrList
         ;

EdgeReciever : NodeId
             | SubGraph
             ;

EdgeRHS : EdgeOp EdgeReciever EdgeRHS
        | EdgeOp EdgeReciever
        ;

EdgeOp : ARROW         // only valid for digraph
       | DDASH         // only valid for graph
       ;

SubGraph : GraphBody
         | SUBGRAPH GraphBody
         | SUBGRAPH ID GraphBody
         ;
```

#### A LL(1) Grammar

A grammar suitable for top down (recursive descent parsing).

This should be LL(1) but I haven't taken the time to prove it. Use with caution.

```
Graphs : Graph Graphs'
       ;

Graphs' : Graph Graphs'
        | e
        ;

Graph : GraphStmt
      | COMMENT
      ;

GraphStmt : GraphType GraphBody
          | GraphType ID GraphBody
          | STRICT GraphType GraphBody
          | STRICT GraphType ID GraphBody
          ;

GraphType : GRAPH
          | DIGRAPH
          ;

GraphBody : LCURLY StmtList RCURLY
          | LCURLY RCURLY
          ;

StmtList : Stmt StmtList'
         | Stmt SEMI StmtList'

StmtList' : Stmt StmtList'
          | Stmt SEMI StmtList'
          | e
          ;

Stmt : ID EQUAL ID
     | NodeStmt
     | EdgeStmt
     | AttrStmt
     | SubGraph
     | COMMENT
     ;

NodeStmt : NodeId
         | NodeId AttrLists
         ;

NodeId : ID
       | ID Port
       ;

Port : COLON ID
     | COLON ID COLON ID // where second ID in "n", "ne", "e", "se", "s", "sw",
                         //                    "w", "nw", "c", "_"
     ;

AttrStmt : AttrType AttrLists
         ;

AttrType : NODE
         | EDGE
         | GRAPH
         ;

AttrLists : AttrList AttrLists'
          ;

AttrLists' : AttrList AttrLists'
           | e
           ;

AttrList : LSQUARE AttrExprs RSQUARE
         | LSQUARE RSQUARE
         ;

AttrExprs : AttrExpr AttrExprs'
          ;

AttrExprs' : AttrExpr AttrExprs'
           | e
           ;

AttrExpr : ID EQUAL ID
         | ID EQUAL ID COMMA
         | ID EQUAL ID SEMI
         ;

EdgeStmt : EdgeReciever EdgeRHS
         | EdgeReciever EdgeRHS AttrList
         ;

EdgeReciever : NodeId
             | SubGraph
             ;

EdgeRHS : EdgeOp EdgeReciever EdgeRHS'
        ;

EdgeRHS' : EdgeOp EdgeReciever EdgeRHS'
         | e
         ;

EdgeOp : ARROW         // only valid for digraph
       | DDASH         // only valid for graph
       ;

SubGraph : GraphBody
         | SUBGRAPH GraphBody
         | SUBGRAPH ID GraphBody
         ;
```


