import requests
a = 0
while a<100:
    a+=1
    response = requests.get("http://127.0.0.1:8080/"+str(a))
    print(response.text)
    