import re

with open('handlers/forum/forumTopicPage.go', 'r') as f:
    content = f.read()

# 1. Extract struct definitions
struct_pattern = re.compile(
    r'\ttype threadWithLabels struct \{.*?\n\t\}\n\n\ttype Data struct \{.*?\n\t\}\n',
    re.DOTALL
)

match = struct_pattern.search(content)
if not match:
    print("Structs not found")
    exit(1)

struct_str = match.group(0)
content = content.replace(struct_str, '')

# Replace threadWithLabels with ThreadWithLabels
struct_str = struct_str.replace('type threadWithLabels', 'type ThreadWithLabels')
struct_str = struct_str.replace('type Data', 'type TopicsPageData')
struct_str = struct_str.replace('*threadWithLabels', '*ThreadWithLabels')
struct_str = struct_str.replace('*Data', '*TopicsPageData')

# Also in the file replace Data with TopicsPageData and threadWithLabels with ThreadWithLabels
content = content.replace('Data{', 'TopicsPageData{')
content = content.replace('*Data', '*TopicsPageData')
content = content.replace('threadWithLabels', 'ThreadWithLabels')

# Insert the structs before TopicsPageWithBasePath
parts = content.split('func TopicsPageWithBasePath')
# Note: struct_str has leading tab/indent, we should unindent it
struct_str = '\n'.join([line[1:] if line.startswith('\t') else line for line in struct_str.split('\n')])

new_content = parts[0] + struct_str + '\nfunc TopicsPageWithBasePath' + parts[1]

with open('handlers/forum/forumTopicPage.go', 'w') as f:
    f.write(new_content)

print("Structs hoisted successfully")
