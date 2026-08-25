import sys

with open("handlers/forum/comments/matchers.go", "r") as f:
    content = f.read()

bad = "topicID := cd.CurrentTopicID()"
good = """topicID := int32(0)
				if topic, _ := cd.CurrentTopic(); topic != nil {
				    topicID = topic.Idforumtopic
				}"""

content = content.replace(bad, good)
with open("handlers/forum/comments/matchers.go", "w") as f:
    f.write(content)
