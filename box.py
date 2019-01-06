#!/usr/bin/python3

import os
import sys
import requests
import json
import hashlib
import base64
requests.packages.urllib3.disable_warnings()

token = sys.argv[1]
file_path = sys.argv[2]
file_name = os.path.split(file_path)[1]
url_get = "https://api.box.com/2.0/folders/0"
url_check = "https://api.box.com/2.0/files/content"
url_post = "https://upload.box.com/api/2.0/files/content"
url_upload = "https://upload.box.com/api/2.0/files/upload_sessions"
header = {"Authorization": "Bearer {0}".format(token)}
header_user = {"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0"}
#proxy = {'http': '127.0.0.1:8888', 'https': '127.0.0.1:8888'}
proxy = {}
minchunked_size = 20000000

def folder_get():
    session = requests.session()
    session.proxies = proxy
    headers = dict(header)
    headers.update(header_user)
    r = session.get(url_get, headers = header)
#    print(r.json())

def file_check(file_name):
    session = requests.session()
    session.proxies = proxy
    headers = dict(header)
    headers.update(header_user)
    data = json.dumps({'name': file_name, 'parent': {'id': '0'}})
    r = session.options(url_check, verify = False, headers = headers, data = data, proxies = proxy)
#    print("Check Result {0}".format(r.status_code))
    return r.status_code == requests.codes.ok

def upload_post(file_path, file_name):
    session = requests.session()
    session.proxies = proxy
    headers = dict(header)
    headers.update(header_user)
    files = {'file': ('unused', open(file_path, 'rb'))}
    data = {'attributes': json.dumps({'name': file_name, 'parent': {'id': '0'}})}
    r = session.post(url_post, verify = False, headers = headers, data = data, files = files)
    return r.status_code
    #print(r.json())

def upload_part(file_path, file_size, file_name):
    session = requests.session()
    session.proxies = proxy
    headers = dict(header)
    headers.update(header_user)
    headers.update({'Content-Type': 'application/json'})
    data = json.dumps({'folder_id': '0', 'file_size': file_size, 'file_name': file_name})
    r = session.post(url_upload, verify = False, headers = headers, data = data)
    s = r.json()
#    print("File Part Result {0}".format(r.status_code))
#    print("Total Part Num {0}, Part Size {1} bytes".format(s['total_parts'], s['part_size']))
    f = open(file_path, 'rb')
    part_size = s['part_size']
    start = 0
    part_num = 0

    while (start < file_size):
        part_num += 1
        end = start + part_size
        if end > file_size:
            end = file_size
        part_hash = hashlib.sha1()
        f.seek(start)
        data = f.read(end - start)
        part_hash.update(data)
        position = end - 1
        hash1 = part_hash.digest()
        b64str = base64.b64encode(hash1).strip().decode('ascii')
        
        header1 = dict(header)
        header1.update(header_user)
        header1.update({"content-range": "bytes {0}-{1}/{2}".format(start, position, file_size)})
        header1.update({"content-type": "application/octet-stream"})
        header1.update({"digest": "sha={0}".format(b64str)})
        url_uploadpart = s['session_endpoints']['upload_part']
        rr = session.put(url_uploadpart, verify = False, headers = header1, data = data)
#        print("Upload Part {0} Result {1}".format(part_num, rr.status_code))
        start += part_size

    url_list = s['session_endpoints']['list_parts']
    header2 = dict(header)
    header2.update(header_user)
    rl = session.get(url_list, verify = False, headers = header2)
    l = rl.json()

    f.seek(0)
    data = f.read()
    file_hash = hashlib.sha1()
    file_hash.update(data)
    b64filestr = base64.b64encode(file_hash.digest()).strip().decode('ascii')
    header3 = dict(header)
    header3.update(header_user)
    header3.update({"content-type": "application/json"})
    header3.update({"digest": "sha={0}".format(b64filestr)})
    data2 = json.dumps({"parts": l['entries']})
    url_commit = s['session_endpoints']['commit']
    rc = session.post(url_commit, verify = False, headers = header3, data = data2)
    #print(rc.json())
    f.close()
    return rc.status_code

def upload_file(file_path):
    file_size = os.path.getsize(file_path)
    if file_size < minchunked_size:
        return upload_post(file_path, file_name)
    else:
        return upload_part(file_path, file_size, file_name)

def main():
    if file_check(file_name):
        if (upload_file(file_path) == 201):
            print("Finished")
        else:
            print("Failed")
    else:
	    print("File check error!")
    #folder_get()
    os._exit(0)

if __name__ == '__main__':
    main()

