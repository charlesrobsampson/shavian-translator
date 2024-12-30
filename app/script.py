import copy
import json
import socket
import os
import sys
import cmudict
import re
import ebooklib
# import epitran
# import espeak
# from ollama import Client
# from ollama import ChatResponse
from ebooklib import epub
from bs4 import BeautifulSoup
# from espeak_phonemizer import Phonemizer

# phonemizer = Phonemizer(default_voice='en-us')

host = 'localhost'
port = 8080
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect((host, port))

# ollamaClient = Client(
#   host='http://ollama:11434'
# )

with open('heteronyms.json', 'r') as file:
    heteronymData = file.read()
heteronyms = json.loads(heteronymData)

cmu = cmudict.dict()

savedWords = {}
# speak = espeak.init()
# epi = epitran.Epitran('eng-Latn')

def process_txt(input_file):
    with open(f"input/{input_file}.txt", "r") as infile, open(f"output/{input_file}-processed.txt", "w") as outfile:
        for line in infile:
            # data = send_to_go(line)
            # outfile.write(data)

            outfile.write(translate(line))
            # processed_content.append(transposed_word)
            # # Send data to Go binary
            # sock.sendall(line.encode())
            # data = sock.recv(1024).decode()
            # print(f"Received from Go binary: {data}")
            # outfile.write(data + "\n")

        print('read file complete')


    sock.close()

def process_epub(input_path):
    print("let's process us an epub")
    book = epub.read_epub(f"input/{input_path}.epub")
    print("got the book")
    new_book = epub.EpubBook()
    print("made new book")

    new_book.set_identifier(book.get_metadata('DC', 'identifier')[0][0])
    new_book.set_title(translate(book.get_metadata('DC', 'title')[0][0]))
    # new_book.set_title(book.get_metadata('DC', 'title')[0][0])
    new_book.set_language(book.get_metadata('DC', 'language')[0][0])

    for creator in book.get_metadata('DC', 'creator'):
        new_book.add_metadata('DC', 'creator', creator[0])

    stylesheets = []
    for styles in book.get_items_of_type(ebooklib.ITEM_STYLE):
        stylesheets.append(styles.get_name())

    for item in book.items:
        if item.get_type() == ebooklib.ITEM_DOCUMENT:
            # Parse content with BeautifulSoup
            original_content = item.get_content().decode('utf-8')
            soup = BeautifulSoup(item.get_content(), 'html.parser')

            for style in stylesheets:
                soup.head.append(BeautifulSoup(f"<link rel=\"stylesheet\" type=\"text/css\" href=\"{style}\"/>"))

            # Preserve XML declaration if it exists
            xml_declaration = ""
            if original_content.startswith("<?xml"):
                xml_declaration_match = re.match(r"(<\\?xml.*?\\?>)", original_content, re.IGNORECASE)
                if xml_declaration_match:
                    xml_declaration = xml_declaration_match.group(1)

            # Transform text nodes in <body>
            for element in soup.body.find_all(string=True):
                parent = element.parent
                # Skip scripts, styles, and other non-visible elements
                if parent.name not in ['script', 'style']:
                    # Apply ROT13 only to text content
                    element.replace_with(translate(element))

            # Reconstruct the document with the preserved XML declaration
            transformed_content = str(soup)
            if xml_declaration:
                transformed_content = xml_declaration + "\n" + transformed_content
            # Create a new item with transformed content
            new_item = epub.EpubItem(
                uid=item.id,
                file_name=item.file_name,
                media_type=item.media_type,
                content=str(soup)
            )
            new_book.add_item(new_item)
        else:
            # Copy other items (e.g., images, stylesheets) as-is
            new_book.add_item(item)

    new_book.toc = book.toc
    new_book.spine = book.spine
    new_book.add_item(epub.EpubNcx())
    new_book.add_item(epub.EpubNav())
    epub.write_epub(f"output/{input_file}-processed.epub", new_book)

def send_to_go(word, request):
    sock.sendall(request)
    
    response = b''
    while b'\n' not in response:
        chunk = sock.recv(4096)
        if not chunk:
            break
        response += chunk

    return json.loads(response).get("transposed", word)

def translate(content: str) -> str:
    # Split the content into words based on non-alphabetic characters
    words = re.findall(r"[a-zA-Z]+(?:['’][a-zA-Z]+)?", content)

    contextWindow = 15
    
    # Process words
    # print(f"we got {len(words)} words")
    for i, word in enumerate(words):
        if word.lower() in savedWords:
            print("found word in saved words: ", word.lower())
            transcribed = savedWords[word.lower()]
        else:
            wordsCopy  = copy.deepcopy(words)
            wordsCopy[i] = f"<{word}>"
            half = contextWindow // 2
            windowStart = i - half
            if windowStart < 0:
                windowStart = 0
            windowEnd = i + half
            if windowEnd >= len(words):
                windowEnd = len(words) - 1
            # print("word index: ", i)
            pronunciation, alphabet, saveWord = get_pronunciation(word, wordsCopy[windowStart:windowEnd])
            request = json.dumps({
                "content": word,
                "pronunciation": pronunciation,
                "alphabet": alphabet,
            }).encode('utf-8') + b'\n'
            transcribed = send_to_go(word, request)
            # print(f"transcribed: {transcribed}")
            # UNCOMMENT
            # if saveWord:
            #     print(f"saving word: {word.lower()} = {transcribed}")
            #     savedWords[word.lower()] = transcribed

        content = content.replace(word, transcribed, 1)
    
    return content

def get_pronunciation(word, context):
    # print(f"getting pronunciation for: {word}")
    word = word.replace('’', "'")
    # UNCOMMENT
    if word.lower() in heteronyms:
        # REMOVE
        return [], "ipa", False
        # # UNCOMMENT
        # # print(f"found heteronym: {word.lower()}")
        # prompt = f"given the context `{' '.join(context)}`, which of the following definitions best fits the usage of <thisWord> in the brackets from the context? `{heteronyms[word.lower()]}` Only return the full json object of the correct definition"
        # # print(prompt)
        # response: ChatResponse = ollamaClient.chat(model='llama3.2', messages=[
        #     {
        #         'role': 'user',
        #         'content': prompt,
        #     },
        # ])
        # # print(f"got the response for {word.lower()}")
        # pronunciationDef = extract_json_from_string(response['message']['content'].replace("'", '"'))
        # if len(pronunciationDef) > 0:
        #     return list(pronunciationDef[0].get("pronunciation")), "ipa", False
        # else:
        #     print(f"weird heteronym response for {word.lower()}:\n{response['message']['content']}")


    pros = cmu.get(word.lower())
    if pros is not None and len(pros) > 0:
        # print(f"found: {pros[0]}")
        return pros[0], "arpabet", False
    else:
        # REMOVE
        return [], "ipa", False
        # UNCOMMENT
        # # print("I must consult the oracle")
        # prompt = f"How would you pronounce `{word.lower()}` using IPA? Only return the IPA like this: `/ipa/`"
        # # format = '{"pronunciation": "<ipaCharacters>"}'
        # # prompt = f"Analyze the following word and give the best guess at its pronunciation in English. The word is probably made up. Please use the IPA to describe its pronunciation and respond in json format like `{format}`. The word is `{word}`. The context is `{' '.join(context)}`"
        # # print(prompt)
        # response: ChatResponse = ollamaClient.chat(model='llama3.2', messages=[
        #     {
        #         'role': 'user',
        #         'content': prompt,
        #     },
        # ])
        # print(f"got the guess for {word.lower()}")
        # print(response['message']['content'])
        # pronunciationDef = extract_ipa_from_string(response['message']['content'])
        # # print("extracted guess: ", pronunciationDef)
        # print("cleaned: ", clean_ipa(pronunciationDef))
        # return list(clean_ipa(pronunciationDef)), "ipa", True

def extract_json_from_string(string):
    """Extracts JSON objects from a string."""
    json_objects = []
    for match in re.finditer(r'{.*?}', string, re.DOTALL):
        try:
            json_objects.append(json.loads(match.group()))
        except json.JSONDecodeError:
            pass  # Ignore invalid JSON
    return json_objects

def extract_ipa_from_string(string):
    match = re.search(r'/.*?/', string)
    if match:
        return match.group()

def clean_ipa(ipa):
    remove = ["/", ";", ":", "'", "ː", ".", "ˈ", "ˌ", "ʔ", "-", "_", "ʰ"]
    for r in remove:
        ipa = ipa.replace(r, "")

    return ipa

if __name__ == "__main__":
    input_file = os.getenv("INPUT_FILE")
    input_type = os.getenv("INPUT_FORMAT")
    print("INPUT_FILE", input_file)
    print("INPUT_FORMAT", input_type)
    if input_type == "txt":
        print('processing txt file', input_file)
        process_txt(input_file)
    elif input_type == "epub":
        print('processing epub file', input_file)
        process_epub(input_file)
    else:
        print("Unsupported file format")
        sys.exit(1)

    sock.close()

    # Process the file
    # process_file("input/input.txt", "output/output.txt")
