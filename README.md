This is an edgestore API. The goal is to give a powerful & faster querying capability to applications based on GAE.

## Bulk Importing Edges

- Define the datastore entity models in `cli/models.go`
- Each entity can output one or many CSV edges
- Every entity should have a separate output folder
- Run the script
- Load the CSV files to `{name}_import` table without any primary key constraints
- Copy from `{name}_import` to `{name}` using `INSERT FROM ... SELECT`
