package main

import (
    "fmt"
    "net/http"
	"io"
    "io/ioutil"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"mime/multipart"
	"encoding/base64"
	"crypto/sha1"
	"crypto/tls"
	)
	
//change source code request.go at line 463: const defaultUserAgent = ""
var url_get string = "https://api.box.com/2.0/folders/0"
var url_check string = "https://api.box.com/2.0/files/content"
var url_post string = "https://upload.box.com/api/2.0/files/content"
var url_upload string = "https://upload.box.com/api/2.0/files/upload_sessions"

var token string
var file_path string
var file_name string
var file_size int64
var minchunked_size int64 = 20000000

var transCfg = &http.Transport{
    Proxy: http.ProxyFromEnvironment,
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
}
var client = &http.Client{Transport: transCfg}

var header_self = map[string][]string{
    "User-Agent": {"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0"},
	"Accept-Encoding": {"gzip, deflate"},
	"Accept": {"*/*"},
	"Connection": {"keep-alive"}}

type parent_struct struct {
    Id string `json:"id"`
}

type options_struct struct {
    Name string `json:"name"`
	Parent parent_struct `json:"parent"`
}

type upload_struct struct {
    Folder_id string `json:"folder_id"`
	File_size int64 `json:"file_size"`
	File_name string `json:"file_name"`
}

type session_struct struct {
    Abort string `json:"abort"`
	Commit string `json:"commit"`
	List_parts string `json:"list_parts"`
	Log_event string `json:"log_event"`
	Status string `json:"status"`
	Upload_part string `json:"upload_part"`
}

type resp_struct struct {
    Id string `json:"id"`
	Num_parts_processed int `json:"num_parts_processed"`
	Part_size int64 `json:"part_size"`
	Session_endpoints session_struct `json:"session_endpoints"`
	Session_expires_at string `json:"session_expires_at"`
	Total_parts int `json:"total_parts"`
	Type string `json:"type"`
}

type part_struct struct {
    Offset int `json:"offset"`
	Part_id string `json:"part_id"`
	Sha1 string `json:"sha1"`
	Size int64 `json:"size"`
}

type resp2_struct struct {
    Entries []part_struct `json:"entries"`
	Limit int `json:"limit"`
	Offset int `json:offset`
	Total_count int `json:"total_count"`
}

type commit_struct struct {
    Parts []part_struct `json:"parts"`
}

func http_get() (int){
    req, err := http.NewRequest("GET", url_get, nil)
	for key, value := range header_self {
	    req.Header[key] = value
	}
	req.Header.Add("Authorization", "Bearer " + token)
//	fmt.Println(req)

	resp, err := client.Do(req)
	    if err != nil {
        fmt.Println("Func 1 error 1")
		return -1
    }
	
    defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	
    if err != nil {
        fmt.Println("Func 1 error 2")
		return -1
    } else {
	    fmt.Println(resp.Status)
	    fmt.Println(string(body))
		return resp.StatusCode
	}
	
	return -1
}

func file_check() (int){
	data := options_struct{Name: file_name, Parent: parent_struct{Id: "0"}}
	option_data, _ := json.Marshal(data)
    req, err := http.NewRequest("OPTIONS", url_check, bytes.NewReader([]byte(option_data)))
	for key, value := range header_self {
	    req.Header[key] = value
	}
	req.Header.Add("Authorization", "Bearer " + token)
//	fmt.Println(req)

	resp, err := client.Do(req)
	    if err != nil {
        fmt.Println("Func 2 error 1")
		return -1
    }
	
	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
	
    if err != nil {
        fmt.Println("Func 2 error 2")
		return -1
    } else {
//	    fmt.Println(resp.Status)
//	    fmt.Println(string(body))
		return resp.StatusCode
	}
	
	return -1
}

func upload_post() int {
	file_read, err := os.Open(file_path)
	if err != nil {
	    fmt.Println("Func 3 error 1")
		return -1
	}
	
	buf := new(bytes.Buffer)
	write := multipart.NewWriter(buf)
	data := options_struct{Name: file_name, Parent: parent_struct{Id: "0"}}
	post_data, _ := json.Marshal(data)
	write.WriteField("attributes", string(post_data)) //In Turn!
	file_write, err := write.CreateFormFile("file", "unused")
	if err != nil {
	    fmt.Println("Func 3 error 2")
		return -1
	}
	
    defer file_read.Close()
	io.Copy(file_write, file_read)
	write.Close()

    req, _ := http.NewRequest("POST", url_post, buf)
    if err != nil {
	    fmt.Println("Func 3 error 3")
		return -1
	}
	for key, value := range header_self {
	    req.Header[key] = value
	}
	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("Content-Type", write.FormDataContentType()) //Post Form	
	//fmt.Println(req)

	resp, err := client.Do(req)
	if err != nil {
        fmt.Println("Func 3 error 4")
		return -1
    }
	
	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	fmt.Println(resp.Status)
//	fmt.Println(string(body))
	
	return resp.StatusCode
}

func upload_part() int {
	data := upload_struct{Folder_id: "0", File_size: file_size, File_name: file_name}
	upload_data, _ := json.Marshal(data)
    req, err := http.NewRequest("POST", url_upload, bytes.NewReader([]byte(upload_data)))
	for key, value := range header_self {
	    req.Header[key] = value
	}
	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("Content-Type", "application/json")
//	fmt.Println(req)
	
	resp, err := client.Do(req)
	    if err != nil {
        fmt.Println("Func 4 error 1")
		return -1
    }
	
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println(resp.Status)
//	fmt.Println(string(body))
	
	var resp_s resp_struct
	err = json.Unmarshal(body, &resp_s)
    if err != nil {
        fmt.Println("Func 4 error 2")
		return -1
    }
	
	part_size := resp_s.Part_size
	url_uploadpart := resp_s.Session_endpoints.Upload_part
	url_list := resp_s.Session_endpoints.List_parts
	file_r, err := os.Open(file_path)
	if err != nil {
        fmt.Println("Func 4 error 3")
		return -1
    }
	defer file_r.Close()
	
	var start, end, position int64 = 0, 0, 0
	part_num := 0
	
	for start < file_size {
	    part_num += 1
		end = start + part_size
		if end > file_size {
		    end = file_size
		}
		position = end - 1
		
		r_buf := make([]byte, end - start)
		r, _ := file_r.ReadAt(r_buf, start)
		if r != int(end - start) {
		    fmt.Println("Func 4 error 4")
			return -1
		}
		
		hash_bytes := sha1.Sum(r_buf)
        base64_str := base64.StdEncoding.EncodeToString(hash_bytes[:]) //Slice!
        req2, err := http.NewRequest("PUT", url_uploadpart, bytes.NewReader([]byte(r_buf)))				
	    for key, value := range header_self {
	        req2.Header[key] = value
	    }
	    req2.Header.Add("Authorization", "Bearer " + token)
	    req2.Header.Add("Content-Type", "application/octet-stream")	
	    req2.Header.Add("content-range", fmt.Sprintf("bytes %d-%d/%d", start, position, file_size))
	    req2.Header.Add("digest", "sha=" + base64_str)
//		fmt.Println(req2)

	    resp2, err := client.Do(req2)
	    if err != nil {
            fmt.Println("Func 4 error 5")
		    return -1
        }		
	    
		defer resp2.Body.Close()
		start += part_size
//	    body2, err := ioutil.ReadAll(resp2.Body)	
//	    fmt.Println(resp2.Status)
//	    fmt.Println(string(body2))
	}
	
    req3, err := http.NewRequest("GET", url_list, nil)			
	for key, value := range header_self {
	    req3.Header[key] = value
	}
	req3.Header.Add("Authorization", "Bearer " + token)
	
	resp3, err := client.Do(req3)
	if err != nil {
        fmt.Println("Func 4 error 6")
	    return -1
    }

    defer resp3.Body.Close()
	body3, _ := ioutil.ReadAll(resp3.Body)	
//	fmt.Println(resp3.Status)
//	fmt.Println(string(body3))
	
	var resp_s2 resp2_struct
	err = json.Unmarshal(body3, &resp_s2)
    if err != nil {
        fmt.Println("Func 4 error 7")
		return -1
    }
	
	commit_data := commit_struct{Parts: resp_s2.Entries}
	commit_json, _ := json.Marshal(commit_data)
    url_commit := resp_s.Session_endpoints.Commit

    req4, err := http.NewRequest("POST", url_commit, bytes.NewReader([]byte(commit_json)))
	for key, value := range header_self {
	    req4.Header[key] = value
	}
	req4.Header.Add("Authorization", "Bearer " + token)
	req4.Header.Add("Content-Type", "application/json")	
    
	file_buf := make([]byte, file_size)
	r, _ := file_r.Read(file_buf)
	if r != int(file_size) {
	    fmt.Println("Func 4 error 8")
		return -1
	}
	hash_file := sha1.Sum(file_buf)
    base64_str := base64.StdEncoding.EncodeToString(hash_file[:])
	req4.Header.Add("digest", "sha=" + base64_str)

	resp4, err := client.Do(req4)
	if err != nil {
        fmt.Println("Func 4 error 9")
	    return -1
    }

    defer resp4.Body.Close()
//	body4, err := ioutil.ReadAll(resp4.Body)
//	fmt.Println(resp4.Status)
//	fmt.Println(string(body4))
	
	return resp4.StatusCode
}

func main() {
	args := os.Args
	if (len(args) != 3) {
	    fmt.Println("Input error!")
		return
	}
	
	token = args[1]
	file_path = args[2]
	file_name = filepath.Base(file_path)

	file_info, err := os.Stat(file_path)
	if(os.IsNotExist(err)) {
	    fmt.Println("File not exist!")
		return
	}
	file_size = file_info.Size()

//	http_get()	
    result := file_check()
    if (result == 200) {
		if (file_size < minchunked_size) {
	        result = upload_post()
	    } else {
	        result = upload_part()
	    }
		if (result == 201) {
		    fmt.Println("Finished")
		}
	} else {
	    fmt.Println("Upload check error!")
	}
}