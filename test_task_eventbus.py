import os
file_path = "handlers/externallink/tasks.go"
with open(file_path, "r") as f:
    content = f.read()

print("Any BackgroundTask in tasks.go?", "BackgroundTask" in content)
