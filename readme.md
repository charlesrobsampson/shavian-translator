# shavian-translator

I decided I wanted to learn to read Shavien better so I made this tool to convert e-books in to Shavian.

It's definitly got some issues... turns out books often have made up words. I'm using a dictionary to get the pronunciations and it's missing some words. Also, heteronyms. I tried adding some LLM support to account for those issues but it's not super reliable and I figured if some words are still left in the latin alphabet, I would still survive.

## usage

### requirements

You'll need `go` and `docker` installed. I think that's it. I tried to keep it as simple as possible and the container will install what it needs to get the job done.

One day I might make this as a tool you can install but for now you'll have to clone the repo and run it from there.

### runnin it

Stick the file you want to convert in the `app/input` folder. Right now only `.epub` and `.txt` files are supported.

Either rename the file to `input.whatever` or run change the env var to reflect the name of the file you want to convert.

The defualts are:
 - `INPUT_FILE=input`
 - `INPUT_FORMAT=epub`

 If you wanted to convert a txt file called `my text.txt` you would run:

 ```bash
 INPUT_FILE="my text" INPUT_FORMAT=txt go run .
 ```

 I think you could figure out to run whatever else you wanted with that.

 It will then create a file called `output_local.<INPUT_FORMAT>` in the root of the project.

 # future potential

 This could be a fun tool to use to convert to any other alphabet. Maybe I'll add rune conversions or something

 I guess let me know if you have any ideas

 # 𐑜𐑫𐑛 𐑤𐑳𐑒!