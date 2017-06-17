package main

import (
	"fmt"
    "io/ioutil"
    "os"
    "strings"
    "golang.org/x/text/transform"
    "bufio"
    "golang.org/x/text/encoding/japanese"
    "github.com/ChimeraCoder/anaconda"
    "github.com/PuerkitoBio/goquery"
    "log"
    "time"
    "strconv"
)
const tmp_path = "/var/tmp/"

func tweet(hashtag string, work_time string) {
    fmt.Println("[EXECUTE] tweet")
    const layout = "2006/01/02"
    const ex_layout = "2006-01-02 15:04:05"
    

    anaconda.SetConsumerKey(os.Getenv("CONSUMERKEY"))
    anaconda.SetConsumerSecret(os.Getenv("CONSUMERSECRET"))
    api := anaconda.NewTwitterApi(os.Getenv("TWETTERAPI1"), os.Getenv("TWETTERAPI2"))
    
    text := work_time + " (ex:" + time.Now().Format(ex_layout) + ") #" + hashtag
    tweet, err := api.PostTweet(text, nil)
    if(err != nil){
        log.Fatal(err)
    }
    fmt.Println(tweet.Text)
}
func getBody(filename,url string) {
    
    fmt.Println("[EXECUTE] getBody")
    _, err := os.Stat(tmp_path + filename)
    if err == nil {
        fmt.Println("[NOTICE] file exist")
        return
    }
    fmt.Println("[NOTICE] file not exist")
    doc, err := goquery.NewDocument(url)
    if err != nil {
        fmt.Print("url scarapping failed")
    }
    res, err := doc.Find("body").Html()
    if err != nil {
        fmt.Print("dom get failed")
    }

    ioutil.WriteFile(tmp_path +filename, []byte(res), os.ModePerm)
}
func getText(filename string) string{
    fmt.Println("[EXECUTE] getText")
    
    fileInfos, _ := ioutil.ReadFile(tmp_path + filename)
    stringReader := strings.NewReader(string(fileInfos))

    utfBody := transform.NewReader(bufio.NewReader(stringReader), japanese.ShiftJIS.NewDecoder())

    doc, err := goquery.NewDocumentFromReader(utfBody)
    if err != nil {
        fmt.Print("url scarapping failed")
    }

    // 出力用配列
    output := make(map[int]map[string]string)
    var i int
    var counter int
    var room_number string
    doc.Find("div #tab_room_all > div.bukken_info > div.bukkeninfo_box03 > table.cost3 > tbody > tr > td").Each(func(_ int, s *goquery.Selection) {
        text := s.Text()
        sjis_text, _ := utf8_to_sjis(text)
        i += 1
        if _, err := strconv.Atoi(sjis_text); err == nil {
            i = 1
        }
        // 初期化
        if i == 1 {
            room_number = sjis_text
        } else if i == 3 {
            output[counter] = make(map[string]string)
            output[counter]["info"] = room_number + "(" + sjis_text + ")"
            counter += 1
        }
    })
    counter = 0
    doc.Find("div #tab_room_all > div.bukken_info > div.bukkeninfo_box01").Each(func(_ int, s *goquery.Selection) {
        
        var sjis_text string

        // 入居可能日
        text := s.Text()
        // 無駄な文字削除
        text = strings.Replace(text, " ", "", -1)
        text = strings.Replace(text, "\n", "", -1)
        // 文字スライス
        text_s := strings.LastIndex(text, "201")
        // 201がある(退室予定がある場合)
        if text_s > 0 {
            text_e := len(text)
            sjis_text = text[text_s:text_e]
        } else {
            sjis_text, err = utf8_to_sjis(text)
        }
        if err != nil {
        //    panic(err)
        }
        output[counter]["date"] = sjis_text
        counter += 1
    })

    // ソート
    var word string
    for k, _ := range output {
        // 出力順が一意になるようにソート
        word += output[k]["info"] + output[k]["date"] + "\n"
    }
    // 出力

    return word
}

func getArgs() (string, string, bool){
    if overArgs(1) == false {
        return "","", false
    }
    if overArgs(2) == false {
        return os.Args[1],"", false
    }
    return os.Args[1], os.Args[2], true
}
func overArgs(i int) (bool) {
    return len(os.Args) > i
}

func utf8_to_sjis(str string) (string, error) {
    iostr := strings.NewReader(str)
    rio := transform.NewReader(iostr, japanese.ShiftJIS.NewEncoder())
    ret, err := ioutil.ReadAll(rio)
    if err != nil {
        return "", err
    }
    return string(ret), err
}

func main() {
    const layout = "2006010215"
    fmt.Println(os.Args[0])
    url, hashtag, err := getArgs()
    if  err != true {
        fmt.Println("[ERROR] empty args")
        fmt.Println(fmt.Sprintf("url........ %s\nhashtag... %s\n %s", url, hashtag))
        return
    }
    fmt.Println("[EXECUTE]")
    fmt.Println(fmt.Sprintf("url........ %s\nhashtag... %s\n %s", url, hashtag))
    // 最低限のチェック
    if strings.Index(url, "http") < 0 {
        fmt.Println("[ERROR] invalid url not http")
        return
    }

    filename := time.Now().Format(layout) + "_" + hashtag + ".html"

    getBody(filename, url)
    word := getText(filename)
    fmt.Println(word)
    

    if word != "" {
        tweet(hashtag, word)
    } else {
        fmt.Println("[NOTICE] does not exist word")
    }
}
