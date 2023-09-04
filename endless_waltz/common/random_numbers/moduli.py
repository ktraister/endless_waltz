#quick python program to convert ssh moduli files into useable generator/prime combos

with open("outfile", 'r') as file:
    for line in file.readlines():
        moduli=line.split(' ')[6]
        res = int(moduli, 16)
        print(str(line.split(' ')[5]), str(res))
