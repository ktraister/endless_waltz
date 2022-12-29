#primary function to encrypt with pad
def pad_encrypt(STRING, PAD):
    final_list = []
    str_list = list(STRING)
    pad_list = list(PAD)
    item_dex = 0
    for item in str_list:
        #print("Working on ", item)
        item_val = ord(item)
        pad_val  = ord(pad_list[item_dex])
        if item_val <= pad_val:
            item_val = item_val + 255
        #print("Item val: " + str(item_val) + " Item Dex: " + str(item_dex) + " Pad_val: " + str(pad_val) )
        final = item_val - pad_val
        final_list.append(final)
        item_dex = item_dex + 1
    return final_list

#primary function to decrypt with the pad
def pad_decrypt(CRYPT_LIST, PAD):
    final_list = []
    str_list = CRYPT_LIST
    pad_list = list(PAD)
    item_dex = 0
    for item in str_list:
        #print("Working on ", item)
        item_val = int(item)
        pad_val  = ord(pad_list[item_dex])
        if item_val >= pad_val:
            item_val = item_val - 255
        #print("Item val: " + str(item_val) + " Item Dex: " + str(item_dex) + " Pad_val: " + str(pad_val) )
        final = item_val + pad_val
        #print("Final num: ", final) 
        final = chr(item_val + pad_val)
        #print("Final num: ", final) 
        final_list += chr(item_val + pad_val)
        item_dex = item_dex + 1
    return final_list
