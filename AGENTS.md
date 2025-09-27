# xsqlparser Agent Guide

## Overview
xsqlparser is a Go SQL parser library that tokenizes and parses SQL statements into an AST. It supports multiple SQL dialects and provides a comprehensive set of SQL statement types.

## Architecture

### Core Components
- **Tokenizer** (`sqltoken/`): Converts SQL input into tokens
- **Parser** (`parser.go`): Builds AST from tokens using recursive descent
- **AST Nodes** (`sqlast/`): Struct definitions for all SQL statement types
- **Dialects** (`dialect/`): SQL flavor-specific keyword handling

### Key Files
- `parser.go`: Main parsing logic with statement-specific parse functions
- `sqlast/stmt.go`: Statement type definitions (CreateTableStmt, SelectStmt, etc.)
- `sqltoken/tokenizer.go`: Token generation from SQL input
- `dialect/`: MySQL, PostgreSQL, and Generic SQL dialects

## Parser Patterns

### Statement Parsing Flow
1. `ParseStatement()` → identifies statement type from first keyword
2. Routes to specific `parseX()` function (e.g., `parseCreate()`, `parseSelect()`)
3. Each parse function consumes tokens and builds corresponding AST node

### Adding New Statement Types
To add support for a new SQL statement (e.g., `CREATE VIRTUAL TABLE`):

1. **Define AST struct** in `sqlast/stmt.go`:
```go
type CreateVirtualTableStmt struct {
    stmt
    Create    sqltoken.Pos
    Name      *ObjectName
    Using     *Ident
    Arguments []string
}
```

2. **Add interface methods** for the new struct:
- `Pos()`, `End()` for position tracking
- `ToSQLString()` for SQL generation
- `WriteTo()` for custom formatting

3. **Update parseCreate()** in `parser.go`:
```go
// Add detection for new keywords
vtok, _, _ := p.parseKeyword("VIRTUAL")
if ok, _, _ := p.parseKeyword("TABLE"); ok {
    if vtok {
        return p.parseCreateVirtualTable(t)
    }
    return p.parseCreateTable(t)
}
```

4. **Implement parse function**:
```go
func (p *Parser) parseCreateVirtualTable(create *sqltoken.Token) (sqlast.Stmt, error) {
    // Parse components and return new struct
}
```

### Token Handling Patterns
- Use `peekToken()` to look ahead without consuming
- Use `nextToken()`/`mustNextToken()` to consume tokens
- Use `consumeToken()` for expected tokens
- Use `parseKeyword()` for keyword matching
- Handle EOF with `errors.Errorf("nextToken failed: %w", err)`

### Common Parser Methods
- `parseObjectName()`: Parses table/view names with optional schema
- `parseIdentifier()`: Single identifier parsing
- `parseExprList()`: Comma-separated expressions
- `parseColumnNames()`: Comma-separated column identifiers

## Error Handling
- Use `log.Panicln()` for syntax errors (following existing pattern)
- Use `errors.Errorf()` with context for parse errors
- Include position information in error messages when helpful

## Testing Strategy
- Use `cmd/astprinter` for manual testing: `echo "SQL;" | go run cmd/astprinter/main.go`
- Build astprinter: `go build ./cmd/astprinter`
- Test edge cases: empty arguments, missing clauses, malformed syntax
- Verify both parsing and SQL generation round-trip

## Code Generation
- Uses `genmark` tool for interface implementations
- Look for `//go:generate genmark` comments
- Run generators after adding new types

## Dialect Handling
- Keywords vary by dialect (MySQL vs PostgreSQL)
- Check `dialect/keywords.go` for reserved words
- Use appropriate dialect when creating parser

## Common Pitfalls
1. **Token consumption**: Always use `peekToken()` before `nextToken()` to avoid consuming tokens needed by caller
2. **Position tracking**: Include proper `From`/`To` positions in all AST nodes
3. **Error recovery**: Be careful with `prevToken()` - ensure state is consistent
4. **Comma handling**: Test both presence and absence of commas in lists
5. **Parentheses**: Balance opening/closing parens and track positions

## Adding New SQL Features
1. Research the SQL syntax to understand all variations
2. Check existing similar statements for patterns
3. Add comprehensive test cases
4. Update error messages in `parseCreate()` to include new statement types
5. Consider both parsing and SQL generation requirements

## Maintenance Notes
- Generated files have `_gen.go` suffix - don't edit manually
- Test files are in `e2e/` directory with sample SQL files
- The codebase follows Go naming conventions with SQL-specific terms