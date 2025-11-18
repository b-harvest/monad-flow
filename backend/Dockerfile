# 빌드 스테이지 - Node.js 20
FROM node:20 AS build

# 작업 디렉토리 설정
WORKDIR /workspace/backend

# 로컬 소스 코드를 컨테이너로 복사
# 주의: Docker 빌드 컨텍스트가 scraper 디렉토리여야 함
COPY . .

# 의존성 설치 및 빌드
RUN npm ci && \
    npm run build

# 실행 스테이지 - Node.js 20
FROM node:20
WORKDIR /backend

# 타임존 설정
RUN apt-get update && \
    apt-get install -y tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Seoul /etc/localtime && \
    echo "Asia/Seoul" > /etc/timezone && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 빌드 스테이지에서 생성된 파일들 복사
COPY --from=build /workspace/backend/dist ./dist
COPY --from=build /workspace/backend/node_modules ./node_modules
COPY --from=build /workspace/backend/package*.json ./

# 환경 변수를 사용하여 애플리케이션 실행
ENTRYPOINT ["node", "dist/main"]