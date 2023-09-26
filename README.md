# go-fixtureplate

**Tools to generate and inspect IPLD data to assist in testing.**

* [Example](#example)
  * [Generate some UnixFS data](#generate-some-unixfs-data)
  * [Explain a CAR file](#explain-a-car-file)
  * [Explain a specific path through the CAR file](#explain-a-specific-path-through-the-car-file)
  * [Use IPFS Trustless Gateway style queries](#use-ipfs-trustless-gateway-style-queries)
* [Installation](#installation)
* [CLI Usage](#cli-usage)
  * [`explain`](#explain)
  * [`generate`](#generate)
* [Generate spec DSL](#generate-spec-dsl)
* [License](#license)

## Example

### Generate some UnixFS data

```console
$ fixtureplate generate \
  'dir(~5*file:1.0kB,~5*file:~102kB,2*dir{sharded}(~10*file:51kB),file:1.0MB{zero},file:10B,file:20B)'
```

Asks for a UnixFS directory structure with some files and directories, packaged into a well-ordered
CAR file. The directory structure is described as:

```
A directory containing:
  → Approximately 5 files of 1.0 kB
  → Approximately 5 files of approximately 102 kB
  → 2 directories sharded with bitwidth 4 containing:
    → Approximately 10 files of 51 kB
  → A file of 1.0 MB containing just zeros
  → A file of 10 B
  → A file of 20 B
```

### Explain a CAR file

```console
$ fixtureplate explain \
  --car bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi.car
```

```
/ipfs/bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi?dag-scope=all
bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi | Directory | /
bafybeih4rlw42wbsamz3zpodkps23kezqqytytdij6ae6sjhq3qvubqwla | HAMTShard | ↳ /Abscond8
bafkreicp3wpwpwpl6fh2d53uo7tprdkodd2b2a4trmauad6eudfczons3e | RawLeaf   |   ↳ /Abscond8/5 [0:50999] (51,000 B)
bafkreicmstgzmn4nqwvswdqdfvh4vzijme36saambqtsjgljyx4yfrbhhi | RawLeaf   |     /Abscond8/i [0:50999] (51,000 B)
bafkreig2bb4fh6e7p2gtxcuxlwkdkkt6egsbtstrdfnaw2hqb2v2qbnoq4 | RawLeaf   |     /Abscond8/wæterscipe [0:50999] (51,000 B)
bafkreidre3o7jeohihhiv3ddbj6zxxe7ftgbvquv6k6lzxh57ba42aq4ei | RawLeaf   |     /Abscond8/ẓẽṕhýr [0:50999] (51,000 B)
bafkreidmzymddrsn4ttxijrc5rjhfidvg5q5r2xso22ktd7lrzvjiryloe | RawLeaf   |     /Abscond8/Polonius [0:50999] (51,000 B)
bafkreienalqkpdgogmjdw7acylwjixmpub7db3have22rjzgc7jqm7cvya | RawLeaf   |     /Abscond8/eorþwerod [0:50999] (51,000 B)
bafkreihwwqxfejrgwx36vq5ilu2hyn5epgvwvl4shtw6relx4t45msb7ne | RawLeaf   |     /Abscond8/Iago [0:50999] (51,000 B)
bafkreigjovj6estaghq7fz7c4cjwgntcthfcdur6ayotwqxyksnwwfpvoq | RawLeaf   |     /Abscond8/juxtapositionally [0:50999] (51,000 B)
bafkreigvz2koxrdifsaznvugtdchyegcpogvsmhsgo3ccf44s54s4b7uiy | RawLeaf   |     /Abscond8/snozzle [0:50999] (51,000 B)
bafkreidv3vhd4iwgtx3v3mxcz53y2ogrpnu7e4ilxwibqvz3er4bdpukbu | RawLeaf   |   /E [0:19] (20 B)
bafkreiapy6vr7t56t5oir762ul3l4hmdgutdte6b5qvwufs3bpvu44jrly | RawLeaf   |   /Fandango [0:111516] (111,517 B)
bafkreigvuhqznqhoxmtu2amw2imgq5plnfttmncplbjkyjarojvjoiueya | RawLeaf   |   /brackle [0:999] (1,000 B)
bafybeidyhi62ammifxk2mtcqdt5hr3d2a2ikoo7nsfbfvohli2njluk7ya | HAMTShard |   /eorþscūa
bafkreico5f2zayg4nnlkdvykwwodpkywnnlmmdm7vsfwrspq3twcr2s2q4 | RawLeaf   |   ↳ /eorþscūa/Persnickety [0:50999] (51,000 B)
bafkreibssfqg3gf6k5zr4kgc2y64rcilfl3lcuv2brupd3qwhk7bermjra | RawLeaf   |     /eorþscūa/Guildenstern [0:50999] (51,000 B)
bafkreibkpahul4fxq42jzunujvbt3j44t2p5had3nrwibghomgar4q23zy | RawLeaf   |     /eorþscūa/Rigmarole [0:50999] (51,000 B)
bafkreicm2ijz6225pp6mtr3na5umyiynkh6bme7kvjyki5acb4hl5b6zsa | RawLeaf   |     /eorþscūa/Titania [0:50999] (51,000 B)
bafybeidbdgc42ogob2s6jv6oqgmsxvanpvai5yw5lgftpkqbu36oebb7ii | HAMTShard |     /eorþscūa [1E]
bafkreid4dvstcnpqdmouolvcyxajnrsp5ohvrtjjtpemagxrnwgtoa2dse | RawLeaf   |     ↳ /eorþscūa/jinty [0:50999] (51,000 B)
bafkreignyiotu3a5be6mbjq5kw4ghwalk2mmqqp2u4mitevzdp6zdgdyf4 | RawLeaf   |       /eorþscūa/mibber [0:50999] (51,000 B)
bafkreifmf6t3l6gl5id6k3zrdpeidiypv4qoav4gi2fw2kmwdivaxfnkg4 | RawLeaf   |       /eorþscūa/5h4r3d [0:50999] (51,000 B)
bafkreif4q6p544er5krzvcpebqwn4ovnnoagink3bjsfm7veodwbkcr4bi | RawLeaf   |     /eorþscūa/brobdingnagian [0:50999] (51,000 B)
bafkreiecrqy3tdnff5ztffcf2a3t6s3evk5tz7gp7nj5iuswchxgye7bgm | RawLeaf   |   /eorþtæfl [0:101689] (101,690 B)
bafkreifxqp3lpiis4furgfwsf7vnu3ildskg47ucgpvyhzndtdayitvj3q | RawLeaf   |   /flippet [0:999] (1,000 B)
bafkreidws6mg2odvhal7lr52vfagjohxy3ofbang7q5johpvhmnl6zsxcy | RawLeaf   |   /go [0:999] (1,000 B)
bafybeibvjrzbx6m66lyh7vu526jkjcvw3tgssjntylas5ikmxmadbtqwee | File      |   /hēahgerefa [0:999999] (1,000,000 B)
bafkreihqnlkfmwgyj34ctwxiobr42xh2p6rxpxp3lygh5albhsgmylqfki | RawLeaf   |   ↳ /hēahgerefa [0:256143] (256,144 B)
bafkreihqnlkfmwgyj34ctwxiobr42xh2p6rxpxp3lygh5albhsgmylqfki | RawLeaf   |     /hēahgerefa [256144:512287] (256,144 B)
bafkreihqnlkfmwgyj34ctwxiobr42xh2p6rxpxp3lygh5albhsgmylqfki | RawLeaf   |     /hēahgerefa [512288:768431] (256,144 B)
bafkreibzz7ccqrjibdd6ew2iqdyccc4hprasqy22rrkgdgmtt7pxhd7way | RawLeaf   |     /hēahgerefa [768432:999999] (231,568 B)
bafkreibs4zufdog5ptkw5pmnohx7rkjcbrngk5qzra4gys3cvtucua6f6a | RawLeaf   |   /klop [0:9] (10 B)
bafkreieanunxuk4a27rb4q6uksyk6iieycaqiizqxa4nw3jurnamk65ewa | RawLeaf   |   /overenthusiastically [0:109434] (109,435 B)
bafkreidsqiwdgu25qgtmoz3rwa7esikryqpr3wfhg7u4krdwm67wuw47yq | RawLeaf   |   /p [0:95598] (95,599 B)
bafkreifme6jeysy43bb64rcwwcik2nb3p3bvbzy22idg2ujzykx64geuse | RawLeaf   |   /ye [0:86970] (86,971 B)
bafkreiaukjbwfe7nbfuabgkh3hy2eahptmlujxhesyb67cywddbvw4nwea | RawLeaf   |   /ŵăνę [0:999] (1,000 B)
```

### Explain a specific path through the CAR file

```console
$ fixtureplate explain \
  --car bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi.car \
  --path /eorþscūa/Guildenstern
```

```
/ipfs/bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi/eorþscūa/Guildenstern?dag-scope=all
bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi | Directory | /
bafybeidyhi62ammifxk2mtcqdt5hr3d2a2ikoo7nsfbfvohli2njluk7ya | HAMTShard | ↳ /eorþscūa
bafkreibssfqg3gf6k5zr4kgc2y64rcilfl3lcuv2brupd3qwhk7bermjra | RawLeaf   |   ↳ /eorþscūa/Guildenstern [0:50999] (51,000 B)
```

### Use IPFS Trustless Gateway style queries

```console
$ fixtureplate explain \
  --car bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi.car \
  --query '/ipfs/bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi/hēahgerefa?dag-scope=entity&entity-bytes=550000:-200000'
```

```
/ipfs/bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi/hēahgerefa?dag-scope=entity&entity-bytes=550000:-200000
bafybeia7igckzgzxeyh5acp2l4etsa3rk2hf533nlwgeggxx4xsjjnxdwi | Directory | /
bafybeibvjrzbx6m66lyh7vu526jkjcvw3tgssjntylas5ikmxmadbtqwee | File      | ↳ /hēahgerefa [0:999999] (1,000,000 B)
bafybeibvjrzbx6m66lyh7vu526jkjcvw3tgssjntylas5ikmxmadbtqwee | File      |   ↳ /hēahgerefa [0:999999] (1,000,000 B)
bafkreihqnlkfmwgyj34ctwxiobr42xh2p6rxpxp3lygh5albhsgmylqfki | RawLeaf   |     ↳ /hēahgerefa [512288:768431] (256,144 B)
bafkreibzz7ccqrjibdd6ew2iqdyccc4hprasqy22rrkgdgmtt7pxhd7way | RawLeaf   |       /hēahgerefa [768432:999999] (231,568 B)
```

## Installation

To install the latest version of go-fixtureplate, run:

```console
$ go install github.com/ipld/go-fixtureplate/cmd/go-fixtureplate@latest
```

Binaries are directly available for download on the [releases page](https://github.com/ipld/go-fixtureplate/releases).

Alternatively, use go-fixtureplate as a library in your application by importing [`github.com/ipld/go-fixtureplate`](https://pkg.go.dev/github.com/ipld/go-fixtureplate).

## CLI Usage

### `explain`

Explain the IPLD contents and paths in a CAR file.

```console
$ fixtureplate explain --car=<car> \
  [--root=<cid>] \
  [--path=<path>] \
  [--scope=<scope>] \
  [--bytes=<byte range>] \
  [--duplicates] \
  [--full-path=false] \
  [--ignore-missing]
```

Or

```console
$ fixtureplate explain --car=<car> --query=<query> [--full-path=false] [--ignore-missing]
```

Where:

* `--car` specifies the path to a CAR file to inspect.
* `--query` specifies an [IPFS Trustless Gateway](https://specs.ipfs.tech/http-gateways/trustless-gateway/) style query to execute. e.g. `/ipfs/<cid>/<path>?dag-scope=<scope>&entity-bytes=<byte range>`. See [the specification](https://specs.ipfs.tech/http-gateways/trustless-gateway/) for full details. Note though that the query here also includes some elements not normally provided on the query string, such as the `dups=y|n` which is normally in the `Accept` header.
* `--root` specifies a root CID to use, overriding the root CID in the CAR file _or_ the `--query`. If not specified, the root CID in the CAR file or `--query` will be used. This may be useful for cases where you are dealing with a CAR without roots, or you want to start from a sub-DAG in the CAR.
* `--path` (default: `/`) specifies a path through the DAG to follow. If not specified, an implicit path of `/` will be used, which will traverse and explain the entire DAG. This would be equivalent to `--query=/ipfs/<cid>?dag-scope=all`.
* `--scope` (or `--dag-scope`, default: `all`) specifies the scope of the traversal at the terminus of the PATH. See the [IPFS Trustless Gateway](https://specs.ipfs.tech/http-gateways/trustless-gateway/) specification for full details. If not specified, the default scope is `all`. Options include `block`, to halt at the block, and `entity` to halt at the block _or_ sharded entity (directory or file) at the terminus of the path.
* `--bytes` (or `--entity-bytes`) specifies the byte range of the entity to return. See the [IPFS Trustless Gateway](https://specs.ipfs.tech/http-gateways/trustless-gateway/) specification for full details. If not specified, the default is to return the entire entity. Supplying a byte range will implicitly set the scope to `entity`.
* `--duplicates` (or `--dups`, default: `true`) specifies whether to include duplicate blocks in the output. If not specified, the default is to include duplicates.
*  `--full-path` (default: `true`) specifies whether to include the full path in the output. If not specified, the default is to include the full path.
*  `--ignore-missing` (default: `false`) specifies whether to ignore missing blocks. If not specified, the default is to error on missing blocks. Turning this on may be useful to explain partial CAR files, such as those downloaded via the IPFS Trustless Gateway using a path, or scope other than `all`.

### `generate`

Generate IPLD data according to a simple DSL that describes the structure of UnixFS file / directory trees.

```console
$ fixtureplate generate [--seed=<seed>] <spec>
```

Where:

* `--seed` specifies a random seed to use for generating the data. If not specified, a random seed will be `0` which should lead to reproducible results.
* `<spec>` is a UnixFS directory structure specification. See [the specification](#generate-spec-dsl) for full details.

`generate` will construct a UnixFS structure in IPLD blocks and output a CAR file containing the data. The CAR will be properly ordered, have the correct root and the name will be `{root cid}.car`. A textual description of the spec will also be printed to stdout in order to clarify what the request was.

## Generate spec DSL

The generate `<spec>` DSL is a simple way to describe a UnixFS directory structure. It is a string that describes a directory structure, with some additional modifiers to control the size and content of the blocks in the DAG. It is intended to be used via the CLI with the `generate` command, but can also be used via the `github.com/ipld/go-fixtureplate/generator` package programmatically to generate UnixFS data for testing purposes.

The basic DSL uses `file:size` and `dir(...)` to describe a directory structure. For example:

```
dir(file:100B,file:200B)
```

Describes a directory containing two files, one of 100 bytes and one of 200 bytes.

The simplest form is a single file. The `file` descriptor _must_ also have an accompanying size after a `:` character. e.g. `file:200KiB`. Once this file is beyond the size of the default chunker used (splitting at 256,144 bytes), it will yield a multi-block sharded file. Otherwise it will be a single block file.

A directory can contain one or more files, which should be comma separated. The simplest directory has a single file. The `dir` descriptor _must_ be followed by a `(...)` containing the files in the directory.

Directories can also be nested, and inteleaved with files. For example:

```
dir(file:100B,dir(file:200B,file:300B),file:400B)
```

Describes a directory containing two files, one of 100 bytes, one of 400 bytes, and one sub-directory containing two files, one of 200 bytes and one of 300 bytes.

File sizes can also be described as **approximate** using the `~` prefix to the size specifier. This inserts some randomness around a target size. For example: `file:~100B` will generate a file of *approximately* 100 bytes. The actual size will be roughly between 90% and 110% of the target size. This is useful for generating data that is not exactly the same size every time, but is still within a reasonable range.

Both files and directories can have **multipliers**. By prefixing `file` or `dir` with `N*` where `N` is a number, the file or directory will be repeated `N` times. For example:

```
dir(5*file:~100B)
```

Describes a directory containing 5 files of approximately 100 bytes each.

Multipliers can also be **approximate** in the same was as file sizes by using `~N*`. For example:

```
dir(~5*file:~100B,~5*dir(~10*file:50B))
```

Describes a directory containing approximately 5 files of approximately 100 bytes each, and approximately 5 directories each containing approximately 10 files of exactly 50 bytes each.

Directories can also be **sharded**. This is done by adding `{sharded}` after the directory descriptor. For example:

```
dir{sharded}(~100*file:~100B)
```

Describes a directory containing approximately 100 files of approximately 100 bytes each, sharded with a default bitwidth of `4`. The normal bitwidth of a UnixFS sharded directory in production is `8`, which yields a fan-out of up to 256 children of each node. Using a bitwidth of `4` yields a fan-out of 16. This is useful for testing purposes as you can generate more collisions and deeper trees using a smaller number of children in the directory.

Alternative bitwidths can be specified by adding a number after `sharded`, as in `{sharded:N}`. For example:

```
dir{sharded:8}(~100*file:~100B)
```

Describes a directory containing approximately 100 files of approximately 100 bytes each, sharded with a bitwidth of `8`.

Files can be **zeroed** by adding `{zero}` after the file descriptor. This will generate a file of the specified size, but with all bytes set to zero. This is primarily useful in generating **duplicate blocks** in your DAG. A file spanning many blocks with all-zeros will generate duplicate blocks for that file, and multiple files in a DAG with zeros will generate duplicate blocks across multiple files. For example:

```
dir(100*file:1MB{zero})
```

Describes a directory containing approximately 100 files of exactly 1MB each, all of which are zeroed. This will generate a DAG with many duplicate blocks. In practice, with the current defaults of `generate`, this will generate a **5** block DAG, where one of those blocks is used **497** times.

Files and directories can be **named** by adding a `{name:"..."}` after the `file` or `dir` descriptor. Multiples cannot be named, and collisions are the responsibility of the user. A mixture of named and non-named files will result in random names being assigned along with the fixed ones. Caution should be applied. For example:

```
dir(dir{name:"boop"}(file:100B{name:"foo"},file:200B{name:"bar"}))
```

Describes a directory containing a directory named `boop` containing two files, one named `foo` and one named `bar`.

When using the `generate` CLI command, a long-form textual description of the spec will be printed to stdout in order to clarify what the request was. For example:

```
dir(~5*file:1.0kB,~5*file:~102kB,2*dir{sharded}(~10*file:51kB),file:1.0MB{zero},file:10B,file:20B)
```

Is described as:

```
A directory containing:
  → Approximately 5 files of 1.0 kB
  → Approximately 5 files of approximately 102 kB
  → 2 directories sharded with bitwidth 4 containing:
    → Approximately 10 files of 51 kB
  → A file of 1.0 MB containing just zeros
  → A file of 10 B
  → A file of 20 B
```

## License

Apache-2.0/MIT © Protocol Labs