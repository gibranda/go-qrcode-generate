# Introduction

go qrcode generator is a simple application to generate qrcode in png format. Data is stored in a temporary folder to save the qrcode and wrapped in zip

## Package

The packages used are as follows

| Name | Link |
| :--- | :--- |
| Echo | `https://github.com/labstack/echo` |
| go qrcode | `https://github.com/skip2/go-qrcode` |
| svg qrcode | `https://github.com/wamuir/svg-qr-code` |
| exelize | `https://github.com/qax-os/excelize` |
| xid | `https://github.com/rs/xid` |

## Install

    go get install

## Run the app

    go run main.go

## Generate QRCode by Total Parameters

This endpoint is used to download the qrcode based on the desired amount of data parameter

```http
GET /download-png/:total
```

| Parameter | Type | Description |
| :--- | :--- | :--- |
| `total` | `string` | **Required**. Total file to generate |

## Responses

File downloaded under zip

## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xid/master/LICENSE).