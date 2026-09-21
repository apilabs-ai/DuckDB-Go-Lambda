# DuckDB Go Lambda

AWS Lambda custom runtime for executing SQL with DuckDB.

## Runtime versions

- DuckDB 1.4.5 LTS (Andium)
- `github.com/duckdb/duckdb-go/v2` v2.5.6
- Go 1.24

The DuckDB version is intentionally pinned and covered by a regression test.

## Test locally

```sh
go test ./...
```

The tests cover the pinned DuckDB version, aggregations, decimals, null values,
timestamps, and the included CSV dataset.

## Invoke

Pass SQL in the `query` property:

```json
{
  "query": "SELECT COUNT(*) AS records FROM 'https://raw.githubusercontent.com/anonranger/Go-DuckDB-Lambda/main/student-data.csv';"
}
```

Multiple statements can be separated with semicolons. Query output is written
to the Lambda logs and the handler returns `See logs below for output`.

Remote files use DuckDB's `httpfs` extension. Lambda must have network access,
and its `HOME` environment variable should point to writable storage such as
`/tmp` so DuckDB can install extensions.

## Example analytics query

```sql
SELECT "Fiscal Year", Career, "Program Level", Campus
FROM 'https://raw.githubusercontent.com/anonranger/Go-DuckDB-Lambda/main/student-data.csv'
ORDER BY RANDOM()
LIMIT 1;
```
