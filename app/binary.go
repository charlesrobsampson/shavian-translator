package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type (
	TranposeRequest struct {
		Content       string   `json:"content,omitempty"`
		Pronunciation []string `json:"pronunciation,omitempty"`
		Alphabet      string   `json:"alphabet,omitempty"`
	}

	TranposeResponse struct {
		Transposed string `json:"transposed,omitempty"`
		Shavian    string `json:"shavian,omitempty"`
	}
)

var (
	fixedWords = map[string]string{
		"the": "𐑞",
		"to":  "𐑑",
		"and": "𐑯",
		"you": "𐑿",
		"of":  "𐑝",
	}
	// https://en.wikipedia.org/wiki/ARPABET
	arpabetToShavianMap = map[string]string{
		"AA0": "𐑭",
		"AA1": "𐑪",
		"AA2": "𐑭",
		"AE0": "𐑨",
		"AE1": "𐑨",
		"AE2": "𐑨",
		"AH0": "𐑩",
		"AH1": "𐑳",
		"AH2": "𐑩",
		"AO0": "𐑷",
		"AO1": "𐑷",
		"AO2": "𐑷",
		"AW0": "𐑬",
		"AW1": "𐑬",
		"AW2": "𐑬",
		"AY0": "𐑲",
		"AY1": "𐑲",
		"AY2": "𐑲",
		"EH0": "𐑧",
		"EH1": "𐑧",
		"EH2": "𐑧",
		"ER0": "𐑼",
		"ER1": "𐑼",
		"ER2": "𐑼",
		"EY0": "𐑱",
		"EY1": "𐑱",
		"EY2": "𐑱",
		"IH0": "𐑦",
		"IH1": "𐑦",
		"IH2": "𐑦",
		"IY0": "𐑦",
		"IY1": "𐑰",
		"IY2": "𐑰",
		"OW0": "𐑴",
		"OW1": "𐑴",
		"OW2": "𐑴",
		"OY0": "𐑶",
		"OY1": "𐑶",
		"OY2": "𐑶",
		"UH0": "𐑫",
		"UH1": "𐑫",
		"UH2": "𐑫",
		"UW0": "𐑵",
		"UW1": "𐑵",
		"UW2": "𐑵",
		"ZH":  "𐑠",
		"Z":   "𐑟",
		"Y":   "𐑘",
		"W":   "𐑢",
		"V":   "𐑝",
		"TH":  "𐑔",
		"T":   "𐑑",
		"SH":  "𐑖",
		"S":   "𐑕",
		"R":   "𐑮",
		"P":   "𐑐",
		"N":   "𐑯",
		"M":   "𐑥",
		"EL":  "𐑤",
		"L":   "𐑤",
		"K":   "𐑒",
		"JH":  "𐑡",
		"HH":  "𐑣",
		"NG":  "𐑙",
		"G":   "𐑜",
		"F":   "𐑓",
		"DH":  "𐑞",
		"D":   "𐑛",
		"CH":  "𐑗",
		"B":   "𐑚",
	}

	ipaDbls = map[string]string{
		"tʃ": "𐑗",
		"dʒ": "𐑡",
		"eɪ": "𐑱",
		"æɪ": "𐑱",
		"aɪ": "𐑲",
		"oʊ": "𐑴",
		"əʊ": "𐑴",
		"aʊ": "𐑬",
		"æw": "𐑬",
		"ɔɪ": "𐑶",
	}

	sngls = map[string]string{
		"p":  "𐑐",
		"b":  "𐑚",
		"t":  "𐑑",
		"d":  "𐑛",
		"d̰": "𐑛",
		"k":  "𐑒",
		"g":  "𐑜",
		"ɡ":  "𐑜",
		"f":  "𐑓",
		"v":  "𐑝",
		"θ":  "𐑔",
		"ð":  "𐑞",
		"s":  "𐑕",
		"z":  "𐑟",
		"ʃ":  "𐑖",
		"ʒ":  "𐑠",
		"j":  "𐑘",
		"w":  "𐑢",
		"ŋ":  "𐑙",
		"h":  "𐑣",
		"l":  "𐑤",
		"ɹ":  "𐑮",
		"ʁ":  "𐑮",
		"ɾ":  "𐑮",
		"r":  "𐑮",
		"m":  "𐑥",
		"n":  "𐑯",
		"ɪ":  "𐑦",
		"i":  "𐑰",
		"ɛ":  "𐑧",
		"e":  "𐑧",
		"æ":  "𐑨",
		"ə":  "𐑩", // pretty much matches a
		"ɜ":  "𐑩",
		"a":  "𐑳", // pretty much matches ə
		"ʌ":  "𐑳", // pretty much matches ə
		"ɔ":  "𐑪", // pretty much same as ɑ
		"ʊ":  "𐑫",
		"u":  "𐑵",
		"ɑ":  "𐑭", // pretty much same as ɔ
		"ɒ":  "𐑷",
	}

	compounds = map[string]string{
		"𐑭𐑮": "𐑸",
		"𐑪𐑮": "𐑸",
		"𐑷𐑮": "𐑹",
		"𐑩𐑮": "𐑹",
		"𐑧𐑮": "𐑺",
		"𐑦𐑮": "𐑽",
		"𐑰𐑮": "𐑽",
		"𐑦𐑩": "𐑾",
		"𐑰𐑩": "𐑾",
		"𐑘𐑵": "𐑿",
	}
)

func main() {
	// Start a TCP server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Go server listening on port 8080...")
	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	processFromSocket(conn)
	fmt.Println("processed content successfully!")
	// // Read data from the Python script
	// scanner := bufio.NewScanner(conn)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	fmt.Printf("Received: %s\n", line)

	// 	// Process data and send it back
	// 	processed := fmt.Sprintf("Processed: %s", line)
	// 	_, err := conn.Write([]byte(processed + "\n"))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
}

func processFromSocket(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req TranposeRequest
		if err := decoder.Decode(&req); err != nil {
			fmt.Printf("Error decoding request: %v\n", err)
			return
		}

		// content := req.Content
		// fmt.Printf("Received content: %s\n", content)

		// transposed := transcribe()

		resp := TranposeResponse{Transposed: transcribe(req)}
		if err := encoder.Encode(resp); err != nil {
			fmt.Printf("Error encoding response: %v\n", err)
			return
		}
	}
}

// func translate(content string) string {
// 	words := strings.FieldsFunc(content, func(r rune) bool {
// 		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
// 	})

// 	// Process words
// 	for _, word := range words {
// 		transcribed := transcribe(word)
// 		content = strings.Replace(content, word, transcribed, 1)
// 	}
// 	return content
// }

func transcribe(req TranposeRequest) string {
	word := strings.ToLower(req.Content)
	// Example transcribe logic
	// fmt.Printf("Transposing word: %s\n", word)
	if fixedWords[word] != "" {
		return fixedWords[word]
	}
	if len(req.Pronunciation) == 0 {
		return req.Content
	}
	if req.Alphabet == "arpabet" {
		shavian := arpabetToShavian(req.Pronunciation)
		// fmt.Printf("got: %s\n", shavian)
		return shavian
	} else if req.Alphabet == "ipa" {
		shavian := ipaToShavian(req.Pronunciation)
		// fmt.Printf("got: %s\n", shavian)
		return shavian
	}
	return req.Content
}

// func processContent(in io.Reader, out io.Writer, transcribeFunc func(string) string) error {
// 	scanner := bufio.NewScanner(in)
// 	writer := bufio.NewWriter(out)

// 	for scanner.Scan() {
// 		line := scanner.Text()

// 		// Split line into words and non-words
// 		words := strings.FieldsFunc(line, func(r rune) bool {
// 			return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
// 		})

// 		// Process words
// 		for _, word := range words {
// 			transcribed := transcribeFunc(word)
// 			line = strings.Replace(line, word, transcribed, 1)
// 		}

// 		// Write processed line
// 		if _, err := writer.WriteString(line + "\n"); err != nil {
// 			return err
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return err
// 	}

// 	return writer.Flush()
// }

func arpabetToShavian(word []string) string {
	shavian := ""
	for _, char := range word {
		if shavianChar, ok := arpabetToShavianMap[char]; ok {
			shavian += string(shavianChar)
		} else {
			shavian += string(char)
		}
	}
	for pair, compound := range compounds {
		shavian = strings.ReplaceAll(shavian, pair, compound)
	}
	return shavian
}

func ipaToShavian(word []string) string {
	shavian := strings.Join(word, "")
	for dbl, shaw := range ipaDbls {
		shavian = strings.ReplaceAll(shavian, dbl, shaw)
	}
	for sngl, shaw := range sngls {
		shavian = strings.ReplaceAll(shavian, sngl, shaw)
	}
	for pair, compound := range compounds {
		shavian = strings.ReplaceAll(shavian, pair, compound)
	}
	return shavian
}
