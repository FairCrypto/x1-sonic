# X1 

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Building the source

Building X1 requires both a Go (version 1.21 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run:

```shell
make all
```
The build output are ```build/x1``` and ```build/x1tool``` executables.

## Initialization of the X1 Database

You will need a genesis file to join a network.

```shell
x1tool --datadir=<target DB path> genesis <path to the genesis file>
```

## Running `x1`

Going through all the possible command line flags is out of scope here,
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `x1` instance.

### Launching a network

Launching `x1` readonly (non-validator) node for network specified by the genesis file:

```shell
x1 --datadir=<DB path>
```

### Configuration

As an alternative to passing the numerous flags to the `x1` binary, you can also pass a
configuration file via:

```shell
x1 --datadir=<DB path> --config /path/to/your/config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export the default configuration:

```shell
x1tool --datadir=<DB path> dumpconfig
```

### Validator

New validator private key may be created with `x1tool --datadir=<DB path> validator new` command.

To launch a validator, you have to use `--validator.id` and `--validator.pubkey` flags to enable 
events emitter. Check the [X1 Documentation](https://docs.xen.network) for the detailed process 
of obtaining the validator ID and registering your initial stake.

```shell
x1 --datadir=<DB path> --validator.id=YOUR_ID --validator.pubkey=0xYOUR_PUBKEY
```

`x1` will prompt you for a password to decrypt your validator private key. Optionally, you can
specify password with a file using `--validator.password` flag.

#### Participation in discovery

Optionally you can specify your public IP to straighten connectivity of the network.
Ensure your TCP/UDP p2p port (5050 by default) isn't blocked by your firewall.

```shell
x1 --datadir=<DB path> --nat=extip:1.2.3.4
```

