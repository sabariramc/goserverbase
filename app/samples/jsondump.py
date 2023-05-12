import json
fp = open("schema.json")
data = json.load(fp)
print(json.dumps(data))