# 임시 폴더 생성 및 이동
mkdir -p /tmp/frida_setup && cd /tmp/frida_setup

# 1. 최신 버전 Frida DevKit 다운로드 (Linux x64 기준)
echo "[*] Downloading Frida DevKit..."
wget -q --show-progress -O frida-devkit.tar.xz https://github.com/frida/frida/releases/download/16.5.6/frida-core-devkit-16.5.6-linux-x86_64.tar.xz

# 2. 압축 해제
echo "[*] Extracting..."
tar -xf frida-devkit.tar.xz

# 3. 시스템 경로로 파일 이동 (sudo 필요)
echo "[*] Installing to /usr/local/..."
sudo mv frida-core.h /usr/local/include/
sudo mv libfrida-core.a /usr/local/lib/

# 4. 정리
cd ~
rm -rf /tmp/frida_setup

echo "[SUCCESS] Frida DevKit installed successfully!"