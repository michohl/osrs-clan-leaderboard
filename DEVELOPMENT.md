# How to Regenerate Database Schemas

If you manually `ALTER` the database schema and want to update the definitions in `jet_schemas` you will need to use the following command:

```shell
go install
go install github.com/go-jet/jet/v2/cmd/jet@latest
jet -source=sqlite -dsn="$YOUR_DB_FILE_PATH" -path=./jet_schemas
```
