## random_numbers
This code is used to generate new `rn.go` files. "All" you have to do is:
```
./new_moduli.sh
```

Remember that both client/server must have the same moduil/generator
pairs available to them to establish a one-time pad. Replace the `rn.go` file 
with care!

### Processing Power
The `ssh-keygen` command used within the script is a single-threaded application
First candidates must be generated, then sifted. This process is extremely 
computationally intensive!
```
(ins)[random_numbers][29_moduli_files]> ssh-keygen -M screen -f moduli-8704.candidates moduli-8704
Wed Nov  8 12:44:16 2023 Found 45 safe primes of 854020 candidates in 175377 seconds
``` 

Primes in the length of 8192, 8448, and 8704 bits were used to generate the 
current `rn.go`. This resulted in 115 safe primes. Generating these primes on my 
desktop with an AMD Ryzen 5 3600 took about a week, with somewhat-concurrent 
processing. 
