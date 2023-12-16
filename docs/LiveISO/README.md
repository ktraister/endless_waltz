# LiveISO
Installation media and live iso instructions
"Installation Media is not a bug, it's a feature"

## Existing ISO
The commands here allow us to update the existing iso with a new build of the
messenger binary

Download the existing iso from S3. Use `Cubic` to start editing.
Filename, VolumeID should be set to ew_messenger_$VERSION.iso
Release should be $VERSION
Disk Name should be ew_messenger
release URL https://www.endlesswaltz.xyz/Downloads#liveISO

install ew_messenger and helpers
```
curl https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/nix_ew_messenger.tar.xz --output nix.tar.xz && tar -xf nix.tar.xz 
rm $files
```

Proceed with the defaults. Generating the new ISO will depend on the
computing power of the host machine. Once complete, you will have the
option (requirement) of testing before finishing generating the ISO.

## New Setup
The Commands in this section should only need to be run once. Afterwards,
we can just re-use the existing LiveISO!

Let's start out with the live iso `linux-lite-6.4-64bit.iso` to be nice 
and light

use the following command to see installed packages and sizes.
find additional candidates for removal for `apt purge`
```
dpkg-query -W -f='${Installed-Size;8}  ${Package}\n' | sort -n
```

purge packages
```
apt purge thunderbird libreoffice* bluez* cups* google-chrome-stable openjdk* samba* gcc-12 cpp-12 gimp* cpp-9 vlc*
```

install ew_messenger and helpers
```
curl https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/nix_ew_messenger.tar.xz --output nix.tar.xz && tar -xf nix.tar.xz 
apt install gnome-startup-applications
```

desktop fixups
```
mkdir -p /home/linux/{.config,Desktop}
ln -s /home/linux/ew_messenger /usr/local/bin/ew_messenger
rm /boot/*.img
echo "Edit the damn grub configs"
```
