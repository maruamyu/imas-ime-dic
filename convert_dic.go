package main

import (
	"archive/zip"
	"bufio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
	"strings"
)

type imeDicEntry struct {
	Yomi, Kanji, Kind, Caption string
}

func (entry *imeDicEntry) isEmpty() bool {
	return (entry.Yomi == "" || entry.Kanji == "")
}

func (entry *imeDicEntry) setFromRow(row []string) {
	if len(row) > 0 {
		entry.Yomi = row[0]
	}
	if len(row) > 1 {
		entry.Kanji = row[1]
	}
	if len(row) > 2 {
		entry.Kind = row[2]
	}
	if len(row) > 3 {
		entry.Caption = row[3]
	}
}

func getUtf16LEBufScanner(fp *os.File) *bufio.Scanner {
	utf16converter := unicode.BOMOverride(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder())
	return bufio.NewScanner(transform.NewReader(fp, utf16converter))
}

func readImeDicFile(srcFile string) (entries []*imeDicEntry, err error) {
	entries = make([]*imeDicEntry, 0)

	var fpInput *os.File
	fpInput, err = os.OpenFile(srcFile, os.O_RDONLY, 0444)
	if err != nil {
		return
	}
	defer fpInput.Close()

	scanner := getUtf16LEBufScanner(fpInput)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 {
			continue
		}
		if strings.HasPrefix(line, "!") {
			continue
		}
		row := strings.Split(line, "\t")
		entry := imeDicEntry{}
		entry.setFromRow(row)
		if entry.isEmpty() {
			continue
		}
		entries = append(entries, &entry)
	}
	err = scanner.Err()
	return
}

func fileTruncateOnCurrentPos(fp *os.File) error {
	pos, err := fp.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := fp.Truncate(pos); err != nil {
		return err
	}
	return nil
}

func createGboardDic(dstFile string, imeDicEntries []*imeDicEntry) error {
	fpZip, errOnOpen := os.OpenFile(dstFile, os.O_CREATE|os.O_RDWR, 0666)
	if errOnOpen != nil {
		return errOnOpen
	}
	defer fpZip.Close()

	zipWriter := zip.NewWriter(fpZip)

	writer, errOnCreate := zipWriter.Create("dictionary.txt")
	if errOnCreate != nil {
		return errOnCreate
	}

	header := []byte("# Gboard Dictionary version:1\n")
	if _, err := writer.Write(header); err != nil {
		return err
	}

	lang := "ja-JP"
	for _, entry := range imeDicEntries {
		line := []byte(entry.Yomi + "\t" + entry.Kanji + "\t" + lang + "\n")
		if _, err := writer.Write(line); err != nil {
			return err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}
	if err := fileTruncateOnCurrentPos(fpZip); err != nil {
		return err
	}
	return nil
}

func createKotoeriDic(dstFile string, imeDicEntries []*imeDicEntry) error {
	fpPlist, errOnOpen := os.OpenFile(dstFile, os.O_CREATE|os.O_RDWR, 0666)
	if errOnOpen != nil {
		return errOnOpen
	}
	defer fpPlist.Close()

	// 単純なXMLなので, xmlパッケージを使わずに直接文字列で出力する
	header := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" + "\n" +
		"<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">" + "\n" +
		"<plist version=\"1.0\">" + "\n" +
		"<array>" + "\n"
	if _, err := fpPlist.WriteString(header); err != nil {
		return err
	}

	for _, entry := range imeDicEntries {
		line := "\t" + "<dict>" + "\n" +
			"\t\t" + "<key>phrase</key>" + "\n" +
			"\t\t" + "<string>" + entry.Kanji + "</string>" + "\n" +
			"\t\t" + "<key>shortcut</key>" + "\n" +
			"\t\t" + "<string>" + entry.Yomi + "</string>" + "\n" +
			"\t" + "</dict>" + "\n"
		if _, err := fpPlist.WriteString(line); err != nil {
			return err
		}
	}

	footer := "</array>" + "\n" + "</plist>" + "\n"
	if _, err := fpPlist.WriteString(footer); err != nil {
		return err
	}

	if err := fileTruncateOnCurrentPos(fpPlist); err != nil {
		return err
	}
	return nil
}

func escapeSlashes(src string) string {
	return strings.ReplaceAll(src, "/", "／")
}

func createSkkDic(dstFile string, imeDicEntries []*imeDicEntry) error {
	fpSkk, errOnOpen := os.OpenFile(dstFile, os.O_CREATE|os.O_RDWR, 0666)
	if errOnOpen != nil {
		return errOnOpen
	}
	defer fpSkk.Close()

	header := ";; imas dic for SKK system" + "\n" +
		";; Keywords: japanese" + "\n" +
		";; okuri-ari entries." + "\n" +
		";; okuri-nasi entries." + "\n"
	if _, err := fpSkk.WriteString(header); err != nil {
		return err
	}

	// よみが同じ単語を一行にまとめる必要がある
	// 順番を保持したいので別途よみのリストを作っておく
	yomiList := make([]string, 0)
	yomiEntriesMap := make(map[string][]*imeDicEntry)
	for _, entry := range imeDicEntries {
		entries, exists := yomiEntriesMap[entry.Yomi]
		if exists == false {
			yomiList = append(yomiList, entry.Yomi)
			entries = make([]*imeDicEntry, 0)
		}
		yomiEntriesMap[entry.Yomi] = append(entries, entry)
	}
	for _, yomi := range yomiList {
		entries, exists := yomiEntriesMap[yomi]
		if exists == false {
			continue
		}
		line := yomi + " "
		for _, entry := range entries {
			line += "/" + escapeSlashes(entry.Kanji)
			if entry.Kind != "" {
				line += ";" + escapeSlashes(entry.Kind)
				if entry.Caption != "" {
					line += "," + escapeSlashes(entry.Caption)
				}
			}
		}
		line += "/" + "\n"
		if _, err := fpSkk.WriteString(line); err != nil {
			return err
		}
	}

	if err := fileTruncateOnCurrentPos(fpSkk); err != nil {
		return err
	}
	return nil
}

func main() {
	srcFile := "dic.txt"
	gboardFile := "dist/gboard.zip"
	kotoeriFile := "dist/macosx.plist"
	skkFile := "dist/skk-jisyo.imas.utf8"

	entries, readErr := readImeDicFile(srcFile)
	if readErr != nil {
		log.Fatal(readErr)
	}

	gbErr := createGboardDic(gboardFile, entries)
	if gbErr != nil {
		log.Fatal(gbErr)
	}

	ktErr := createKotoeriDic(kotoeriFile, entries)
	if ktErr != nil {
		log.Fatal(ktErr)
	}

	skkErr := createSkkDic(skkFile, entries)
	if skkErr != nil {
		log.Fatal(skkErr)
	}
}
