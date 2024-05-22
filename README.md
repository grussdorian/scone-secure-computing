**More information about scone can be found on their [website](https://sconedocs.github.io/)**

## Example 1

### Problem

We want to create a confidential program `P1` that

1. takes all its arguments, writes each arguments as a line in file `/volumes/v1/file`

2. `P1` should run under a policy `N`

3. It asks for an OTP to be able to run

4. `P1` should take the host arguments (i.e., arguments provided by user)

5. Host arguments: specify with `@@1` (first host argument) `@@2` (second host argument), In the policy, command: `print-arg-env arg1 @@1 @@2`
   Mix with arguments from policy: e.g., `arg1` is provided by policy

6. Ensure that file `/volumes/v1` is transparently encrypted

7. Define the volume in a separate policy `V`

8. `V` exports the volume to policy `P1`. Export the volume to namespace `<N>`

### Solution

**All the policies are under the folder `policies`**
**All the logs are under the folder `logs`**
**The encrypted files have been copied under the folder called `encrypted_files`**

First we define some prerequisites and write them in our `.bashrc` file so we don't have to write verbose commands again and again. In the `.bashrc` file:

```bash
..
..

alias scone="docker run $MOUNT_SGXDEVICE  --network=host -it -v `pwd`:/work registry.scontain.com/sconecuratedimages/crosscompilers bash"

function determine_sgx_device {
    export SGXDEVICE="/dev/sgx_enclave"
    export MOUNT_SGXDEVICE="--device=/dev/sgx_enclave"
    if [[ ! -e "$SGXDEVICE" ]] ; then
        export SGXDEVICE="/dev/sgx"
        export MOUNT_SGXDEVICE="--device=/dev/sgx"
        if [[ ! -e "$SGXDEVICE" ]] ; then
            export SGXDEVICE="/dev/isgx"
            export MOUNT_SGXDEVICE="--device=/dev/isgx"
            if [[ ! -c "$SGXDEVICE" ]] ; then
                echo "Warning: No SGX device found! Will run in SIM mode." > /dev/stderr
                export MOUNT_SGXDEVICE=""
                export SGXDEVICE=""
            fi
        fi
    fi
}

determine_sgx_device
```

Then we run the docker container with the correct SGX device

```bash
scone
```

The source file of the C program `P1` used is given here. Since the docker container we have is very minimal, we use the `cat` command to write and read files due to the unavailability of text editors like nano and vim. The working directory is called `/work`

```bash
cd /work
cat > P1.c << EOF
# Then the rest of the c program ...
```

Now we have to compile the program (optimisation level 3 and with the debugging information attached)

```bash
scone-gcc P1.c -g -O3 -o P1
```

Now running the program

```bash
mkdir -p /volumes/v1/
./P1 hello world
```

Running of this program has created the file `file.txt` with unencrypted contents. Output in file `log1`

The output is unencrypted. We want to encrypt it. For that we need to create a scone policy with a session and attest it from CAS.

But first we want to create a namespace called `SESSION_EX1` and then we would create two different policies one for `P1` and another for our volume `V` which run under `SESSION_EX1` namespace.

File `p1_namespace.yaml` defines our namespace. Then we create two different files `policy_for_V.yaml` and `policy_for_p1.yaml` as the names suggest.

To create a policy `N`, we do the following:

```bash
export SESSION_EX1=EX1-$RANDOM-$RANDOM
export SESSION_P1=P1-$RANDOM-$RANDOM
export SESSION_V=V-$RANDOM-$RANDOM
export MRENCLAVE=`SCONE_HASH=1 ./P1 Assignment Hardik Ghoshal 5184856`
```

We are basically running `P1` to find its hash called `MRENCLAVE` that we need to use in our policy to ensure integrity of the code while running.

The content of the file `/volumes/v1/file.txt` is shown in `output1.log` file. The contents are in clear text.

The $RANDOM will generate a random number every time we want to create a session. This means that for every session we create, we get a different name. This is really convenient as to avoid duplicate session names without much hassle.

Now we have to specify where CAS and LAS are running.

```bash
export SCONE_CAS_ADDR=141.76.44.93
export SCONE_LAS_ADDR=141.76.50.190
```

And we have to define our OTPSECRET variable which stores the base 32 encoded value of a string which we arbitrarily chose. Our string `iamtestingsconehello` and our base 32 encoded version of the string is `NFQW25DFON2GS3THONRW63TFNBSWY3DP`

one can easily encode their string into a valid base32 encoding from this [website](https://emn178.github.io/online-tools/base32_encode.html)

```bash
export OTPSECRET=NFQW25DFON2GS3THONRW63TFNBSWY3DP
```

Then we need to attest the CAS session

```bash
scone cas attest $SCONE_CAS_ADDR --only_for_testing-trust-any --only_for_testing-debug  --only_for_testing-ignore-signer -C -G -S
```

Now the running version of CAS is attested.

Now we write the policy files by using the cat command (because nano vim and other text editors are not available)

```bash
cat > policy_for_P1.yaml <<EOF
name: $SESSION_EX1/$SESSION_P1
... rest of the file
EOF
```

Similarly we had to paste all of the policy files

Before running the code we want to create sessions that we previously defined in our policy files.

```bash
export PREDECESSOR_namespace=$(scone session create p1_namespace.yaml)
echo $PREDECESSOR_namespace c960c7e4f28db2c3f9e228fadce615ece8be5ddff242b8b7059c56b6f5339a7a
export PREDECESSOR_V=$(scone session create policy_for_V.yaml)
echo $PREDECESSOR_V 1536e4f1bb9f6f0894ea0c3175524a46302f425b78465cd60a64c3d8477b36f5
export PREDECESSOR_P1=$(scone session create policy_for_P1.yaml)
echo $PREDECESSOR_P1 f3c09d4fcf3d432dfd4c77197def087e64c4a778c16bdde5d7cd48bfbe639d17

```

The predecessor variable is used when we want to chain policies further (for example if we want to decrypt the file we would need the predecessor value)

Finally before running the code, we derive an OTP from [this website](https://totp.info/) where we have to put our `OTPSECRET`

Then

```bash
export OTP=123456 # correct otp from the website
```

Output of running the list directory command before running the code is shown in `output2.log` file.

And then running the program

```bash
SCONE_CONFIG_ID=$SESSION_EX1/$SESSION_P1/P1@$OTP ./P1 Hardik Ghoshal
```

Finally output of running the list directory command after running the code is given in `output3.log`

## Example 2

In this example we would create 3 policy files for three different 32 byte hex keys which scone generates and passes as arguments to our second program P2, written in Go. P2 takes tree keys k1, k2 and k3 and creates a combined key by XORing them.

We create policy names

```bash
export POLICY_K1=K1-$RANDOM-$RANDOM
export POLICY_K2=K2-$RANDOM-$RANDOM
export POLICY_K3=K3-$RANDOM-$RANDOM
export POLICY_P2=P2-$RANDOM-$RANDOM
```

Then we write the policy files using cat as we did before and then we run the policy files.

```bash
export PREDECESSOR_POLICY_K1=$(scone session create policy_for_k1.yaml)
export PREDECESSOR_POLICY_K2=$(scone session create policy_for_k2.yaml)
export PREDECESSOR_POLICY_K2=$(scone session create policy_for_k2.yaml)
```

```bash
echo $PREDECESSOR_POLICY_K1 8ea006363a6ecfc08f71d5e672969f32cc9e62ae522f7915b12b6b6195b09d3a
echo $PREDECESSOR_POLICY_K2
0b2dd7fb4dc7c48995ad52197e9ec84ce23240a605b1f134ead3ff2c717dcabc
echo $PREDECESSOR_POLICY_K3
db50bec81c6d67409a696fef861920af56c52c87d420e530bac2d5164ca89b6b
```

We run the policy for decryption

```bash
export PREDECESSOR_POLICY_P2=$(scone session create policy_for_decryption.yaml)
echo $PREDECESSOR_POLICY_P2
4235d3b8b214a2b67da56eb521ac853dcda300cb4921d6aecdd7bdf9525c87ed
```

Finally the output after running the go program

```bash
SCONE_MODE=HW SCONE_CONFIG_ID=$SESSION_EX1/$POLICY_P2/decrypt ./P2
```

```bash
Combined key: 50416a5537705f476a740f6c734b594e543968567c2e68342e00517f4e306d54

Contents of the file before encryption:

bash-5.1  cat ../volumes/v1/file.txt
argv: ./P1 Assignment Hardik Ghoshal 5184856
environ:
SESSION_V=V-7360-8710
SCONE_LAS_ADDR=141.76.50.190
HOSTNAME=se1-fujitsu
SESSION_P1=P1-19922-25906
REALGCC=/usr/bin/gcc
PWD=/work
HOME=/root
MRENCLAVE=3e5d9ffa08ca89682db603315e7c9a39e95d962044478a8b2f994d003644c41f
SCONE_CAS_ADDR=141.76.44.93
TERM=xterm
OTP=NFQW25DFON2GS3THONRW63TFNBSWY3DP
SHLVL=1
SESSION=P1-25855-16288
PREDECESSOR_V=
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
OTPSECRET=NFQW25DFON2GS3THONRW63TFNBSWY3DP
OLDPWD=/
_=./P1


Encrypted file written to /out/file.aes

Contents of the file after decryption:


bash-5.1  cat ../volumes/v1/file.txt
argv: ./P1 Assignment Hardik Ghoshal 5184856
environ:
SESSION_V=V-7360-8710
SCONE_LAS_ADDR=141.76.50.190
HOSTNAME=se1-fujitsu
SESSION_P1=P1-19922-25906
REALGCC=/usr/bin/gcc
PWD=/work
HOME=/root
MRENCLAVE=3e5d9ffa08ca89682db603315e7c9a39e95d962044478a8b2f994d003644c41f
SCONE_CAS_ADDR=141.76.44.93
TERM=xterm
OTP=NFQW25DFON2GS3THONRW63TFNBSWY3DP
SHLVL=1
SESSION=P1-25855-16288
PREDECESSOR_V=
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
OTPSECRET=NFQW25DFON2GS3THONRW63TFNBSWY3DP
OLDPWD=/
_=./P1
```

Output of the encrypted file `file.aes` is inside `encrypted_files/file.aes`
