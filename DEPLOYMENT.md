# Dokumentasi Deployment ke Google Cloud Run

## Daftar Isi
- [Prerequisites](#prerequisites)
- [Step 1: Clone Repository](#step-1-clone-repository)
- [Step 2: Review dan Update Kode](#step-2-review-dan-update-kode)
- [Step 3: Setup Google Cloud CLI](#step-3-setup-google-cloud-cli)
- [Step 4: Build dan Test Docker Image Lokal](#step-4-build-dan-test-docker-image-lokal)
- [Step 5: Push ke Artifact Registry](#step-5-push-ke-artifact-registry)
- [Step 6: Deploy ke Cloud Run](#step-6-deploy-ke-cloud-run)
- [Step 7: Test Service](#step-7-test-service)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

Pastikan Anda sudah menginstall:
- **Docker Desktop** - untuk build dan test container
- **Google Cloud SDK (gcloud)** - untuk deployment ke GCP
- **Git** - untuk clone repository
- **Akun Google Cloud Platform** dengan billing enabled

---

## Step 1: Clone Repository

Clone repository dari GitHub:

```bash
git clone https://github.com/gibranda/go-qrcode-generate.git
cd go-qrcode-generate
```

---

## Step 2: Review dan Update Kode

### 2.1 Update go.mod

Edit file `go.mod` dan update versi Go ke 1.24:

```go
module go-qrcode

go 1.24
```

### 2.2 Update main.go

Tambahkan logic untuk membaca PORT dari environment variable. Tambahkan di awal fungsi `main()`:

```go
func main() {
	// Create qrcode directory if not exists
	if err := os.MkdirAll("qrcode", 0755); err != nil {
		log.Fatal("Failed to create qrcode directory:", err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/download-png/:total", downloadPNG)
	e.GET("/download-svg", downloadSVG)
	e.GET("/download-excel", getExcelFile)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3010"
	}

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
	
	// ... rest of the code
}
```

### 2.3 Update Dockerfile

Buat/update `Dockerfile` dengan konfigurasi berikut:

```dockerfile
FROM docker.io/golang:1.24-alpine as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o appsvc

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/appsvc .

CMD ["./appsvc"]
```

**Poin penting dalam Dockerfile:**
- `GOOS=linux GOARCH=amd64` - memastikan binary compatible dengan Cloud Run
- Multi-stage build - menghasilkan image yang lebih kecil
- Distroless image - lebih aman dan minimal

---

## Step 3: Setup Google Cloud CLI

### 3.1 Login ke Google Cloud

```bash
gcloud auth login
```

Command ini akan membuka browser untuk login ke akun Google Anda.

### 3.2 Lihat Daftar Project

```bash
gcloud projects list
```

Output contoh:
```
PROJECT_ID        NAME              PROJECT_NUMBER
gibranda-project  Gibranda Project  4692970255
```

### 3.3 Set Active Project

```bash
gcloud config set project PROJECT_ID
```

Ganti `PROJECT_ID` dengan ID project Anda (bukan PROJECT_NUMBER).

### 3.4 Set Default Region

```bash
gcloud config set run/region asia-southeast2
```

Region lain yang bisa digunakan:
- `asia-southeast1` (Singapore)
- `asia-southeast2` (Jakarta)
- `us-central1` (Iowa)
- `europe-west1` (Belgium)

### 3.5 Verifikasi Konfigurasi

```bash
gcloud config list
```

Output contoh:
```
[core]
account = your-email@gmail.com
project = gibranda-project
[run]
region = asia-southeast2
```

### 3.6 Enable Required APIs

```bash
gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
```

---

## Step 4: Build dan Test Docker Image Lokal

### 4.1 Build Docker Image

```bash
docker build --platform linux/amd64 -t go-qrcode-generator:local .
```

**Penting:** Flag `--platform linux/amd64` memastikan image compatible dengan Cloud Run.

### 4.2 Test Container Lokal

```bash
# Run container
docker run -d -p 3010:8080 -e PORT=8080 --name qrcode-test go-qrcode-generator:local

# Wait for container to start
sleep 3

# Test endpoint
curl http://localhost:3010/

# Expected output: Hello, World!
```

### 4.3 Stop dan Hapus Container Test

```bash
docker stop qrcode-test
docker rm qrcode-test
```

---

## Step 5: Push ke Artifact Registry

### 5.1 Configure Docker Authentication

```bash
gcloud auth configure-docker asia-southeast2-docker.pkg.dev
```

Ketik `Y` untuk konfirmasi.

### 5.2 Tag Image untuk Artifact Registry

```bash
docker tag go-qrcode-generator:local \
  asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest
```

Ganti `PROJECT_ID` dengan ID project Anda.

### 5.3 Push Image

```bash
docker push asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest
```

---

## Step 6: Deploy ke Cloud Run

### 6.1 Deploy Service

```bash
gcloud run deploy go-qrcode-generator \
  --image asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest \
  --region asia-southeast2 \
  --allow-unauthenticated \
  --platform managed
```

**Penjelasan flags:**
- `--image` - URL image dari Artifact Registry
- `--region` - Region untuk deploy service
- `--allow-unauthenticated` - Mengizinkan akses public tanpa authentication
- `--platform managed` - Deploy ke Cloud Run (fully managed)

### 6.2 Output Deployment Berhasil

```
✓ Deploying... Done.
  ✓ Creating Revision...
  ✓ Routing traffic...
  ✓ Setting IAM Policy...
Done.
Service [go-qrcode-generator] revision [go-qrcode-generator-00001-xxx] has been deployed and is serving 100 percent of traffic.
Service URL: https://go-qrcode-generator-xxxx-xx.a.run.app
```

**Catat Service URL** untuk testing!

---

## Step 7: Test Service

### 7.1 Test Basic Endpoint

```bash
curl https://YOUR-SERVICE-URL.run.app/
```

Expected output: `Hello, World!`

### 7.2 Test Generate QR Codes (ZIP)

```bash
# Generate 10 QR codes
curl -O https://YOUR-SERVICE-URL.run.app/download-png/10

# File qrcode.zip akan terdownload
```

### 7.3 Test SVG QR Code

```bash
curl https://YOUR-SERVICE-URL.run.app/download-svg
```

### 7.4 Test Excel Download

```bash
curl -O https://YOUR-SERVICE-URL.run.app/download-excel
```

---

## Troubleshooting

### Problem: "exec format error"

**Penyebab:** Binary tidak compatible dengan platform Cloud Run (linux/amd64)

**Solusi:**
1. Pastikan Dockerfile menggunakan `GOOS=linux GOARCH=amd64` saat build
2. Build dengan flag `--platform linux/amd64`

```bash
docker build --platform linux/amd64 -t go-qrcode-generator:local .
```

### Problem: "Container failed to start and listen on port"

**Penyebab:** Aplikasi tidak membaca environment variable `PORT`

**Solusi:** Update main.go untuk membaca `PORT` dari environment variable:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "3010"
}
e.Logger.Fatal(e.Start(":" + port))
```

### Problem: "gcloud: command not found"

**Penyebab:** Google Cloud SDK belum terinstall

**Solusi:** Install Google Cloud SDK:

**MacOS:**
```bash
# Menggunakan Homebrew
brew install --cask google-cloud-sdk

# Atau download dari:
# https://cloud.google.com/sdk/docs/install
```

**Linux:**
```bash
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

### Problem: "Project number instead of project ID"

**Penyebab:** gcloud menggunakan project number, bukan project ID

**Solusi:** 
```bash
# List semua project
gcloud projects list

# Set menggunakan PROJECT_ID (bukan PROJECT_NUMBER)
gcloud config set project PROJECT_ID
```

### Melihat Logs Cloud Run

Jika ada error saat deployment, lihat logs:

```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=go-qrcode-generator" --limit 20
```

---

## Commands Cheat Sheet

```bash
# Build image
docker build --platform linux/amd64 -t go-qrcode-generator:local .

# Test lokal
docker run -d -p 3010:8080 -e PORT=8080 --name qrcode-test go-qrcode-generator:local
curl http://localhost:3010/
docker stop qrcode-test && docker rm qrcode-test

# Configure & push ke Artifact Registry
gcloud auth configure-docker asia-southeast2-docker.pkg.dev
docker tag go-qrcode-generator:local asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest
docker push asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest

# Deploy ke Cloud Run
gcloud run deploy go-qrcode-generator \
  --image asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest \
  --region asia-southeast2 \
  --allow-unauthenticated \
  --platform managed

# View logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=go-qrcode-generator" --limit 20

# List services
gcloud run services list

# Delete service (jika perlu)
gcloud run services delete go-qrcode-generator --region asia-southeast2
```

---

## Update Service (Re-deploy)

Jika Anda melakukan perubahan pada kode dan ingin update service:

```bash
# 1. Rebuild image
docker build --platform linux/amd64 -t asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest .

# 2. Push ke registry
docker push asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest

# 3. Deploy ulang (gcloud akan otomatis update)
gcloud run deploy go-qrcode-generator \
  --image asia-southeast2-docker.pkg.dev/PROJECT_ID/cloud-run-source-deploy/go-qrcode-generator:latest \
  --region asia-southeast2 \
  --allow-unauthenticated \
  --platform managed
```

---

## Monitoring & Management

### View Service Details

```bash
gcloud run services describe go-qrcode-generator --region asia-southeast2
```

### View Revisions

```bash
gcloud run revisions list --service go-qrcode-generator --region asia-southeast2
```

### Set Traffic Split (untuk Blue/Green deployment)

```bash
gcloud run services update-traffic go-qrcode-generator \
  --region asia-southeast2 \
  --to-revisions REVISION-1=50,REVISION-2=50
```

### Update Scaling Configuration

```bash
gcloud run services update go-qrcode-generator \
  --region asia-southeast2 \
  --min-instances 0 \
  --max-instances 10
```

### Update Memory & CPU

```bash
gcloud run services update go-qrcode-generator \
  --region asia-southeast2 \
  --memory 512Mi \
  --cpu 1
```

---

## Cost Optimization Tips

1. **Set min-instances to 0** - Menghindari biaya saat tidak ada traffic
2. **Set max-instances** - Mencegah auto-scaling berlebihan
3. **Use appropriate region** - Pilih region terdekat dengan user
4. **Monitor usage** - Gunakan Cloud Monitoring untuk track usage
5. **Set up billing alerts** - Notifikasi jika biaya melebihi threshold

---

## Security Best Practices

1. **Jangan expose secrets** - Gunakan Secret Manager untuk API keys
2. **Use authentication** jika service tidak public:
   ```bash
   gcloud run deploy SERVICE --no-allow-unauthenticated
   ```
3. **Regularly update dependencies** - Scan vulnerabilities dengan `docker scan`
4. **Use least privilege IAM** - Berikan minimal permission yang diperlukan
5. **Enable Cloud Armor** - Untuk DDoS protection jika perlu

---

## Referensi

- [Google Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
- [Cloud Run Pricing](https://cloud.google.com/run/pricing)

---

**Catatan:** Ganti semua `PROJECT_ID` dan `YOUR-SERVICE-URL` dengan nilai actual dari project dan service Anda.

**Dibuat:** 2025-10-22  
**Author:** Deployment Guide untuk Go QR Code Generator
