



ra = "/dev/random"

r1 = open(ra, "r")
val1 = r1.read()


val1 = val1.encode('utf-8').strip()
r1.close()
print(val1)


