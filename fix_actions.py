import re

with open('.github/workflows/ci.yml', 'r') as f:
    content = f.read()

content = content.replace('actions/checkout@v4', 'actions/checkout@v7')
content = content.replace('actions/checkout@v6', 'actions/checkout@v7')
content = content.replace('actions/setup-go@v5', 'actions/setup-go@v7')
content = content.replace('actions/setup-node@v4', 'actions/setup-node@v7')
content = content.replace('goreleaser/goreleaser-action@v6', 'goreleaser/goreleaser-action@v7')

with open('.github/workflows/ci.yml', 'w') as f:
    f.write(content)

print("Done updating actions")
