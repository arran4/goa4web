import re
import sys

def main():
    schema_file = 'database/schema.mysql.sql'

    try:
        with open(schema_file, 'r') as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: {schema_file} not found.")
        sys.exit(1)

    # Regex to find CREATE TABLE statements
    table_regex = re.compile(r'CREATE TABLE (?:IF NOT EXISTS )?`(\w+)` \((.*?)\);', re.DOTALL)

    tables_without_deleted_at = []

    for match in table_regex.finditer(content):
        table_name = match.group(1)
        table_body = match.group(2)

        # Check if deleted_at exists in the table body
        if '`deleted_at`' not in table_body and 'deleted_at' not in table_body:
             # Exclude some tables that likely don't need soft deletes (e.g., join tables, schema_version)
             # This is a heuristic; manual review might be needed.
            if not table_name.startswith('schema_version') and \
               not table_name.startswith('searchwordlist_has_') and \
               not table_name.endswith('_search') and \
               not table_name.endswith('Search') and \
               table_name != 'schema_version':
                tables_without_deleted_at.append(table_name)

    if tables_without_deleted_at:
        print("Tables without deleted_at:")
        for table in tables_without_deleted_at:
            print(table)
    else:
        print("All relevant tables have deleted_at.")

if __name__ == "__main__":
    main()
