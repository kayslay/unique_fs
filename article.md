---
title: Using Hashing to prevent file duplicates and save storage
published: true
description: 
tags: golang,tutorial,programming,learning
cover_image: 
---

This article aims to show a method of preventing duplicate files and saving storage space by using hashing algorithms.

There are cases where a video, image, or media might be trending due to a meme, news, or whatever can make anything trend. In such a scenario, the same file gets uploaded repeatedly by different users. With repeated uploads, the load on the server will increase, and more space used to save duplicates of the same file.

```text
A hash function is any function that can be used to map data of arbitrary size to fixed-size values.
The values returned by a hash function are called hash values, hash codes, digests, or simply hashes. 
The values are used to index a fixed-size table called a hash table

- Wikipedia
```

## How it works

I'm going to use a method that makes use of a file store, a key-value store (`hash-storage linker`), and a `reference object` to a file.

### The File Store
It is responsible for saving the actual file uploaded by users. Storage platforms or services like S3, Google Storage, Block Storage, for example, can be used as a file store.

### Hash Storage Linker

It is a DB that holds the file hash value as a key and the path to the file on the file store (e.g S3) as its value. A KV-store DB is great for storing this type of data.

It serves the middle-man between the file store and the file reference.

```
Note: The Hash-Storage is not required for this to work. If the server uses a single folder to store
 all the files, it can just hash the value of the file, and save the file using the hash value. 

Using the Hash-Storage makes checking the existence of the file faster since it is a KV-store. 

It can also act as an abstraction for separating files from file reference.
using it allows us to use multiple distributed file stores.
```

Provided the files exist, a `hash-storage linker` record must exist.

### The Uploader's File Reference Object

It holds the reference to the hash and the contents of the file's metadata. It can be modified by the uploader, which means it can be created, updated, and deleted by the user.

Now That we have seen the entities that play a role in making this work, the next step is to show how they come together to solve the problem.

## Upload Flow On The Frontend

The first step is for the frontend application (this could be a browser, mobile app, or another client) uploading the file, is to generate the hash value of the file to be uploaded. After generating the hash value check if a file with the hash value exists on the `hash-storage linker`([File exist endpoint](https://documenter.getpostman.com/view/2909688/SzzrXYyE?version=latest#19825fe7-9200-467f-b000-bab5a34ee51d)).

If the file does not exist, the frontend application will send the file's content and the information(metadata) about the file ([Upload file endpoint](https://documenter.getpostman.com/view/2909688/SzzrXYyE?version=latest#5400e91d-d835-42d7-8c25-1517cbc0fbdc)). This info passed with can be the filename, uploader info, time e.t.c. 

If the file exists, The frontend will only send the information needed by the file reference and the hash of the file. It reduces latency and load on the frontend and backend applications.


## Upload Flow On The Backend Server

The backend server waits for a request from a frontend application. The server handles the upload request based on the content passed in the body of the request.

The server expects the file metadata (information), and a `body` or `hash` value. Use `body` when the file does not exist and `hash` when the file exists. 

If the file does not exist, `body` must be sent in the request body so that the server knows it's a new file and creates the file on the file storage service. After creating the file the server creates a record on the `hash-storage linker` links the file path on the file store to `reference objects`.

The last thing the server does is create a reference to the file for the uploader to use. The data contained in the `reference object` contains the hash key value from the `hash-storage linker` and extra information about the uploaded file. Info like time of upload, uploader's id, file_name, access .e.t.c. Just anything the user requires for the application to work.

When using the `hash` in place of `body` in the request body, create only the file's `reference object`.

## Get File Flow On The Backend
When getting a file, the backend server provides two endpoints. One endpoint gets the [file reference object](https://documenter.getpostman.com/view/2909688/SzzrXYyE?version=latest#ae158aaf-410f-4c41-9792-ed709cf6a538),
 and the other one gets the [actual file content](https://documenter.getpostman.com/view/2909688/SzzrXYyE?version=latest#4652202d-9a69-4573-81a6-1b0a81321ca4).

Fetching the file `reference object` is of a less task on the server. 
The hash returned with the file `reference object` can be used to fetch the file, but in the case where the storage provider provides something similar to [S3 Presigned Link](https://medium.com/r/?url=https%3A%2F%2Fdocs.aws.amazon.com%2FAmazonS3%2Flatest%2Fdev%2FPresignedUrlUploadObject.html), the server could return a presigned get link. The reason for this is to reduce the load on your server and move it to S3, which is more likely better equipped for pulling files from the internet (just saying).

## Get File Flow On The Frontend

The frontend makes use of [fetch file reference endpoint](https://documenter.getpostman.com/view/2909688/SzzrXYyE?version=latest#ae158aaf-410f-4c41-9792-ed709cf6a538) to get the file. The hash value extracted from the file `reference object` data helps identify the path to fetch the actual file on the backend or a locally stored data (this depends on the application).

Not all frontend implementation will have to save files locally, but if your app has that ability to save files, then storing it locally will save time when getting a file that has been downloaded.

WhatsApp's status is an example of a case where saving the file on the device is a cool idea. Sometimes a user sees the same video or image multiple times on different people's status.

Instead of downloading the same image multiple times, using the hash to save the file will reduce the number of times we request the server to get the same media file. The only thing that changes is the metadata (caption, uploaded time, user id).

## Deleting Files
Since there is a possibility that different `reference objects` can depend on the same file, deleting it won't be so straight forward.

When Deleting the `reference object`, we need to know if other `reference objects` reference the same `hash-storage linker` as the soon to be deleted ref object. If the no ref object points to the hash-storage link, we remove the file from the file storage platform, and delete the `hash-storage linker` record, and the reference file.

If other ref object points to the same `hash-storage linker` record, then we delete only the ref file and leave the `hash-storage linker` and file store alone.

## How the code will look like

The code in this article will focus on the backend part of the uploader. Postman will be used to represent the frontend service making the request.

Why use postman? It makes me lazy while at the same time doing a good job showing the frontend flow.

[Postman Documentation](https://documenter.getpostman.com/view/2909688/SzzrXYyE)

## Hashing Method

The reason for hashing the file is to create a unique key for a file. So for this, we must use a hashing algorithm that has fewer chances of a collision.

```text
In computer science, a collision or clash is a situation that occurs when two distinct 
pieces of data have the same hash value, checksum, fingerprint, or cryptographic digest. 

- Wikipedia
```

I should not be saying this, but the hashing algorithm on the frontend and backend should be the same.

The hashing algorithm I'll be making use of in the example is md5. It does the job of explaining my point here. For serious projects, use something that has fewer collision conflicts.

Check this [github repo](https://github.com/kayslay/unique_fs) for the implementation. 

## Conclusion

I honestly don't know what to conclude this.

THE END 🤷‍♂. BYE 🤷‍♂.
