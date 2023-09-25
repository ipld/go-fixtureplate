# go-fixtureplate

**Tools to generate and inspect IPLD data to assist in testing.**

## Example

### Generate some UnixFS data

```
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

```
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

```
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

### Use [Trustless Gateway] style queries

```
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

```
go get github.com/ipld/go-fixtureplate
```

## Usage

