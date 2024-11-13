# Build stage
FROM golang:1.23 AS backend-builder

# Set the working directory
WORKDIR /go/src/github.com/h4-poc/service

# Copy go mod and sum files
COPY go.mod go.sum ./

RUN go mod download

# Copy the source code
COPY . .

# Build arguments
ARG VERSION
ARG BUILD_DATE
ARG GIT_COMMIT
ARG INSTALLATION_MANIFESTS_URL="github.com/h4-poc/service/manifests/base"
ARG INSTALLATION_MANIFESTS_THIRD_PARTY="github.com/h4-poc/service/manifests/third-party"

RUN CGO_ENABLED=0 GOOS=linux go build -v -o service \
    -ldflags "-X 'github.com/h4-poc/service/pkg/store.version=${VERSION}' \
    -X 'github.com/h4-poc/service/pkg/store.buildDate=${BUILD_DATE}' \
    -X 'github.com/h4-poc/service/pkg/store.gitCommit=${GIT_COMMIT}' \
    -X 'github.com/h4-poc/service/pkg/store.installationManifestsURL=${INSTALLATION_MANIFESTS_URL}' \
    -X 'github.com/h4-poc/service/pkg/store.installationManifestsThirdParty=${INSTALLATION_MANIFESTS_THIRD_PARTY}' \
    cmd/service/service.go

FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY web/package.json web/package-lock.json* ./
RUN \
  if [ -f package-lock.json ]; then npm ci; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Rebuild the source code only when needed
FROM base AS web-builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web/ .

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
# ENV NEXT_TELEMETRY_DISABLED=1

RUN \
  if [ -f package-lock.json ]; then npm run build; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

# debug tools
RUN apk add --no-cache curl

ENV NODE_ENV=production
# Uncomment the following line in case you want to disable telemetry during runtime.
# ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=web-builder /app/public ./public

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=web-builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=web-builder --chown=nextjs:nodejs /app/.next/static ./.next/static
# Copy the binary from the builder stage
COPY --from=backend-builder /go/src/github.com/h4-poc/service/service .

USER nextjs