#!/usr/bin/env python3
#quick python program to convert ssh moduli files into useable generator/prime combos

print("package main")
print()
print("var moduli_pairs = []string{")

with open("outfile", 'r') as file:
    for line in file.readlines():
        # Time Type Tests Tries Size Generator Modulus
        modType=line.split(' ')[1]
        if modType != "2":
            continue
        moduli=line.split(' ')[6]
        res = int(moduli, 16)
        print("        \"" + str(line.split(' ')[5]) + ":" + str(res) + "\",")

print("}")
