# Go Best Practices Rules

This directory contains best practices documentation for commonly used Go packages. These rules serve as guidelines for writing idiomatic, performant, and maintainable Go code.

## Available Rules

| Rule File | Package | Description |
|-----------|---------|-------------|
| [golang-general.md](golang-general.md) | Go Core | General Go programming best practices |
| [gin.md](gin.md) | gin-gonic/gin | Gin web framework patterns |
| [redis.md](redis.md) | redis/go-redis | Redis client best practices |
| [clickhouse.md](clickhouse.md) | ClickHouse/clickhouse-go | ClickHouse driver patterns |
| [postgresql.md](postgresql.md) | jackc/pgx | PostgreSQL client best practices |
| [cobra.md](cobra.md) | spf13/cobra | CLI application patterns |
| [viper.md](viper.md) | spf13/viper | Configuration management |

## How to Use These Rules

1. **Code Reviews**: Reference these rules during code reviews to ensure consistency
2. **New Projects**: Use as a checklist when setting up new Go projects
3. **Learning**: Study these patterns to understand Go best practices
4. **Documentation**: Link to specific rules in your project documentation

## Rule Format

Each rule file follows this structure:

1. **Overview**: Brief description of the package and its purpose
2. **Installation**: How to add the package to your project
3. **Best Practices**: Core patterns and recommendations
4. **Common Patterns**: Code examples for typical use cases
5. **Anti-Patterns**: What to avoid and why
6. **Performance Tips**: Optimization recommendations
7. **Testing**: How to test code using the package

## Contributing

When adding new rules or updating existing ones:

1. Follow the established format
2. Include working code examples
3. Explain the "why" behind each recommendation
4. Reference official documentation where applicable
5. Test all code examples before committing

## References

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
