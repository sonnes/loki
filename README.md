This is an experimental API that stores & serves edges. The goal is to store different relations - edges, to a relational database and provide faster querying for counts, joins, feeds etc for applications based on GAE.

This is just a learning project. I started to learn building services in Go.

## Bulk Importing Edges

- Define the datastore entity models in `cli/models.go`
- Each entity can output one or many CSV edges
- Every entity should have a separate output folder
- Run the script
- Load the CSV files to `{name}_import` table without any primary key constraints
- Copy from `{name}_import` to `{name}` using `INSERT FROM ... SELECT`
