result = ''
message = ''
choice = ''
otp = "ssbzyztxprxshswqfrzatxpwnfthtcgkvbjtbvhhmyuuqhbqzahsrqgdvzepbmsqgugbqhcagqnqypdxgepxnvahnf"
finstr = ''

#this code is a simple ceaser cipher, but with unicode values for keys
#unfortunately, this doesn't really respect spaces

while choice != '-1':
    choice = input("\nDo you want to encrypt or decrypt the message?\nEnter 1 to Encrypt, 2 to Decrypt, -1 to Exit Program: ")

    if choice == '1':
        message = input("\nEnter the message to encrypt: ")

        for i in range(0, len(message)):
            charnum = ord(message[i])
            #print("\nmessage[i]:", message[i])
            #print("charnum:", charnum)
            keynum  = ord(otp[i])
            #print("\nkeychar[i]", otp[i])
            #print("keynum[i]:", keynum)
            resnum = (charnum + keynum) % 128
            #print("\nresnum:", resnum)
            reschar = chr(resnum)
            #print("reschar:", reschar)
            finstr = finstr + reschar

        print("Final String:", finstr)
        #print (result + '\n\n')
        #result = ''

    elif choice == '2':
        message = input("\nEnter the message to decrypt: ")
        for i in range(0, len(message)):
            charnum = ord(message[i])
            #print("\nmessage[i]:", message[i])
            #print("charnum:", charnum)
            keynum  = ord(otp[i])
            #print("\nkeychar[i]", otp[i])
            #print("keynum[i]:", keynum)
            resnum = (charnum - keynum) % 128
            #print("\nresnum:", resnum)
            reschar = chr(resnum)
            #print("reschar:", reschar)
            finstr = finstr + reschar

        print("Final String:", finstr)



    elif choice != '-1':
        print ("You have entered an invalid choice. Please try again.\n\n")
