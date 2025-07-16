# 🚀 CI/CD Pipeline Documentation

This document explains how to set up and configure the CI/CD pipeline for Rawboard.

## 📋 Pipeline Overview

The GitHub Actions workflow provides:

- **Continuous Integration**: Automated testing, linting, and security checks
- **Continuous Deployment**: Automatic Docker image builds and pushes to Docker Hub
- **Quality Assurance**: Go formatting, vulnerability scanning, and test coverage

## 🔧 Pipeline Jobs

### 1. **Test Job** (`test`)

Runs on every push and pull request:

- ✅ Go code formatting check
- ✅ Static analysis with `go vet`
- ✅ Vulnerability scanning with `govulncheck`
- ✅ Unit tests with race detection and coverage
- ✅ Database integration tests (with Redis)

### 2. **Docker Job** (`docker`)

Runs only on pushes to `main` branch after tests pass:

- 🐳 Builds Docker image
- 🏷️ Tags with multiple versions
- 📦 Pushes to Docker Hub repository: `2ryan09/rawboard`

## 🔑 Required GitHub Secrets

To enable the full CI/CD pipeline, you **must** configure these secrets in your GitHub repository.

### Setting Up Secrets

1. **Navigate to Repository Settings**

   ```
   GitHub Repository → Settings → Secrets and variables → Actions
   ```

2. **Click "New repository secret"** for each required secret below

### Required Secrets

| Secret Name       | Description                      | Where to Get It                       | Example Format                   |
| ----------------- | -------------------------------- | ------------------------------------- | -------------------------------- |
| `DOCKER_USERNAME` | Your Docker Hub username         | Docker Hub account                    | `2ryan09`                        |
| `DOCKER_PASSWORD` | Docker Hub Personal Access Token | Docker Hub → Security → Access Tokens | `dckr_pat_xxxxxxxxxxxxxxxxxxxxx` |

## 🐳 Docker Hub Setup

### Step 1: Create Personal Access Token

1. **Log into Docker Hub** (hub.docker.com)
2. **Go to Account Settings** → **Security** → **Personal Access Tokens**
3. **Click "New Access Token"**
   - **Token description**: `GitHub Actions - Rawboard`
   - **Access permissions**: `Public Repo Read/Write`
4. **Generate and copy the token** ⚠️ _You'll only see it once!_

### Step 2: Create Docker Hub Repository

Ensure you have a repository named `rawboard` in your Docker Hub account:

- Repository URL: `https://hub.docker.com/r/2ryan09/rawboard`
- Visibility: Public (recommended) or Private

## 🏷️ Docker Image Tags

The pipeline automatically creates multiple tags for better version management:

| Tag Format   | When Created         | Example                         | Use Case                 |
| ------------ | -------------------- | ------------------------------- | ------------------------ |
| `latest`     | Every push to `main` | `2ryan09/rawboard:latest`       | Production deployments   |
| `main`       | Every push to `main` | `2ryan09/rawboard:main`         | Branch-specific          |
| `main-<sha>` | Every push to `main` | `2ryan09/rawboard:main-a1b2c3d` | Specific commit tracking |

## 🚦 Pipeline Triggers

### When Tests Run:

- ✅ Push to `main` branch
- ✅ Push to `develop` branch
- ✅ Pull requests to `main` branch

### When Docker Images Are Built:

- ✅ Push to `main` branch **only** (after tests pass)
- ❌ Pull requests (security - no images built)
- ❌ Push to other branches

## 📊 Pipeline Status

You can monitor pipeline status:

1. **GitHub Actions Tab**: See real-time build progress
2. **Repository Badge**: Add to README for status visibility
3. **Docker Hub**: View pushed images and download stats

## 🔍 Troubleshooting

### Common Issues

**❌ Docker login failed**

```
Error: Error response from daemon: unauthorized
```

**Solution**: Check `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets are set correctly

**❌ Docker push denied**

```
Error: denied: requested access to the resource is denied
```

**Solution**: Verify Docker Hub repository exists and PAT has write permissions

**❌ Tests failing**

```
Error: Database connection failed
```

**Solution**: Check if Redis service is properly configured in workflow

### Secret Validation

To verify secrets are working:

1. **Check Actions logs** for successful Docker login
2. **Look for** this message in logs:
   ```
   Login Succeeded
   ```
3. **Verify** images appear in Docker Hub after successful push

## 🛡️ Security Best Practices

### Docker Hub PAT Security:

- ✅ Use Personal Access Tokens (not passwords)
- ✅ Scope tokens to minimum required permissions
- ✅ Rotate tokens regularly (every 6-12 months)
- ✅ Monitor token usage in Docker Hub logs

### GitHub Secrets Security:

- ✅ Never log or echo secret values
- ✅ Use repository secrets (not environment secrets) for this project
- ✅ Review secret access periodically

## 📈 Usage After Setup

Once configured, the pipeline works automatically:

### For Development:

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes and push
git push origin feature/new-feature

# Create pull request → Tests run automatically
```

### For Production:

```bash
# Merge to main branch → Tests + Docker build + Push
git checkout main
git merge feature/new-feature
git push origin main
```

### Using the Docker Image:

```bash
# Pull latest version
docker pull 2ryan09/rawboard:latest

# Run the container
docker run -d -p 8080:8080 \
  -e RAWBOARD_API_KEY=your-key \
  -e VALKEY_URI=your-redis-uri \
  2ryan09/rawboard:latest
```

## 🚀 Next Steps

After setting up secrets:

1. **Push to main** to trigger first Docker build
2. **Check Docker Hub** for successful image push
3. **Test pulling** the image locally
4. **Add CI/CD badge** to main README (optional)

## 📞 Support

If you encounter issues:

1. **Check GitHub Actions logs** for detailed error messages
2. **Verify all secrets** are correctly configured
3. **Ensure Docker Hub repository** exists and is accessible
4. **Review this documentation** for setup steps

---

**✅ Ready to deploy!** Once secrets are configured, every push to `main` will automatically build and deploy your Docker image.
