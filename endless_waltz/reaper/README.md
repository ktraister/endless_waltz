This component is used to gather random data from background radiation and store it in s3 for pad use 

Quickstart Guide for SDR:
https://www.rtl-sdr.com/rtl-sdr-quick-start-guide/

HOW TO ACTUALLY CAPTURE DATA:
https://www.aaronscher.com/wireless_com_SDR/RTL_SDR_AM_spectrum_demod.html

command to use for capture:
rtl_sdr -f 130122000 -g 40 -s 25000000 -n 2500000000000000000000000000000000000 - | xxd -b

#RTL ENTROPY
https://www.rtl-sdr.com/rtl-sdr-as-a-hardware-random-number-generator-with-rtl_entropy/
https://pthree.org/2015/06/16/hardware-rng-through-an-rtl-sdr-dongle/

Now i have a fork of the RTL-entropy software to turn my SDR into an RNG. 

TEMP rancher pass: EpjAVeyzpGenotSx

