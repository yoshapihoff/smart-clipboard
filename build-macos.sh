#!/bin/bash

# Скрипт для сборки macOS приложения SmartClipboard
# Поддержка macOS 11+ (Big Sur и новее)
# Использование: ./build-macos.sh [arm64|x86_64|universal]

set -e  # Выход при ошибке

# Константы
APP_NAME="SmartClipboard"
BINARY_NAME="smart-clipboard"
BUILD_DIR="builds/darwin"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
ICON_FILE="assets/icon.png"

# Определение архитектуры для сборки
ARCH=${1:-"universal"}
MACOS_MIN_VERSION="11.0"
SDK_PATH="$(xcrun --show-sdk-path --sdk macosx)"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go не установлен. Пожалуйста, установите Go."
        exit 1
    fi
    
    if ! command -v xcrun &> /dev/null; then
        log_error "xcrun не найден. Установите Xcode Command Line Tools: xcode-select --install"
        exit 1
    fi
    
    if ! command -v sips &> /dev/null; then
        log_error "sips не найден. Это стандартный инструмент macOS."
        exit 1
    fi
    
    if ! command -v iconutil &> /dev/null; then
        log_error "iconutil не найден. Это стандартный инструмент macOS."
        exit 1
    fi
    
    if ! command -v codesign &> /dev/null; then
        log_error "codesign не найден. Это стандартный инструмент macOS."
        exit 1
    fi
    
    if ! command -v lipo &> /dev/null; then
        log_error "lipo не найден. Это стандартный инструмент macOS."
        exit 1
    fi
    
    # Проверка версии macOS
    MACOS_VERSION=$(sw_vers -productVersion | cut -d. -f1-2)
    REQUIRED_VERSION="11.0"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$MACOS_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        log_warn "Текущая версия macOS: $MACOS_VERSION. Рекомендуется macOS 11.0 или новее для сборки."
    else
        log_info "Версия macOS: $MACOS_VERSION (подходит для сборки)"
    fi
    
    log_info "Все зависимости найдены"
}

# Сборка бинарного файла
build_binary() {
    log_info "Сборка Go бинарного файла для архитектуры: $ARCH..."
    
    # Создание директории для сборки
    mkdir -p "$BUILD_DIR"
    
    # Установка переменных окружения для сборки под macOS 11+
    export MACOSX_DEPLOYMENT_TARGET=$MACOS_MIN_VERSION
    export CGO_ENABLED=1
    export CGO_CFLAGS="-mmacosx-version-min=$MACOS_MIN_VERSION"
    export CGO_LDFLAGS="-mmacosx-version-min=$MACOS_MIN_VERSION"
    
    case $ARCH in
        "arm64")
            log_info "Сборка для Apple Silicon (arm64)"
            GOOS=darwin GOARCH=arm64 go build \
                -ldflags "-s -w" \
                -o "$BUILD_DIR/$BINARY_NAME" ./cmd
            ;;
        "x86_64")
            log_info "Сборка для Intel (x86_64)"
            GOOS=darwin GOARCH=amd64 go build \
                -ldflags "-s -w" \
                -o "$BUILD_DIR/$BINARY_NAME" ./cmd
            ;;
        "universal")
            log_info "Сборка Universal Binary (Apple Silicon + Intel)"
            
            # Временные директории для бинарных файлов
            ARM64_BUILD_DIR="$BUILD_DIR/arm64"
            X86_64_BUILD_DIR="$BUILD_DIR/x86_64"
            
            mkdir -p "$ARM64_BUILD_DIR" "$X86_64_BUILD_DIR"
            
            # Сборка для arm64
            GOOS=darwin GOARCH=arm64 go build \
                -ldflags "-s -w" \
                -o "$ARM64_BUILD_DIR/$BINARY_NAME" ./cmd
            
            # Сборка для x86_64
            GOOS=darwin GOARCH=amd64 go build \
                -ldflags "-s -w" \
                -o "$X86_64_BUILD_DIR/$BINARY_NAME" ./cmd
            
            # Создание универсального бинарного файла
            lipo -create \
                "$ARM64_BUILD_DIR/$BINARY_NAME" \
                "$X86_64_BUILD_DIR/$BINARY_NAME" \
                -output "$BUILD_DIR/$BINARY_NAME"
            
            # Очистка временных директорий
            rm -rf "$ARM64_BUILD_DIR" "$X86_64_BUILD_DIR"
            
            log_info "Universal Binary создан"
            ;;
        *)
            log_error "Неподдерживаемая архитектура: $ARCH. Используйте arm64, x86_64 или universal"
            exit 1
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        log_info "Бинарный файл успешно собран для $ARCH"
    else
        log_error "Ошибка сборки бинарного файла"
        exit 1
    fi
}

# Создание структуры приложения
create_app_structure() {
    log_info "Создание структуры приложения..."
    
    # Проверка и создание директории для сборки
    if [ ! -d "$BUILD_DIR" ]; then
        mkdir -p "$BUILD_DIR"
    fi
    
    # Удаление старой версии если существует
    if [ -d "$APP_BUNDLE" ]; then
        log_info "Удаление старой версии приложения..."
        rm -rf "$APP_BUNDLE"
    fi
    
    # Создание структуры директорий
    log_info "Создание директорий: $APP_BUNDLE/Contents/{MacOS,Resources}"
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    mkdir -p "$APP_BUNDLE/Contents/Resources"
    
    # Проверка существования бинарного файла
    if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
        log_error "Бинарный файл не найден: $BUILD_DIR/$BINARY_NAME"
        exit 1
    fi
    
    # Копирование бинарного файла
    log_info "Копирование бинарного файла..."
    cp "$BUILD_DIR/$BINARY_NAME" "$APP_BUNDLE/Contents/MacOS/"
    
    # Установка прав на выполнение
    chmod +x "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME"
    
    # Проверка создания структуры
    if [ -d "$APP_BUNDLE/Contents/MacOS" ] && [ -d "$APP_BUNDLE/Contents/Resources" ]; then
        log_info "Структура приложения успешно создана"
    else
        log_error "Не удалось создать структуру приложения"
        exit 1
    fi
}

# Создание иконки .icns
create_icon() {
    log_info "Создание иконки приложения..."
    
    if [ ! -f "$ICON_FILE" ]; then
        log_warn "Иконка $ICON_FILE не найдена, пропускаю создание .icns"
        return
    fi
    
    # Временная директория для иконок
    ICONSET_DIR="icon.iconset"
    
    # Удаление старой директории если существует
    if [ -d "$ICONSET_DIR" ]; then
        rm -rf "$ICONSET_DIR"
    fi
    
    mkdir -p "$ICONSET_DIR"
    
    # Создание иконок разных размеров
    sips -z 16 16 "$ICON_FILE" --out "$ICONSET_DIR/icon_16x16.png" 2>/dev/null || true
    sips -z 32 32 "$ICON_FILE" --out "$ICONSET_DIR/icon_16x16@2x.png" 2>/dev/null || true
    sips -z 32 32 "$ICON_FILE" --out "$ICONSET_DIR/icon_32x32.png" 2>/dev/null || true
    sips -z 64 64 "$ICON_FILE" --out "$ICONSET_DIR/icon_32x32@2x.png" 2>/dev/null || true
    sips -z 128 128 "$ICON_FILE" --out "$ICONSET_DIR/icon_128x128.png" 2>/dev/null || true
    sips -z 256 256 "$ICON_FILE" --out "$ICONSET_DIR/icon_128x128@2x.png" 2>/dev/null || true
    sips -z 256 256 "$ICON_FILE" --out "$ICONSET_DIR/icon_256x256.png" 2>/dev/null || true
    sips -z 512 512 "$ICON_FILE" --out "$ICONSET_DIR/icon_256x256@2x.png" 2>/dev/null || true
    sips -z 512 512 "$ICON_FILE" --out "$ICONSET_DIR/icon_512x512.png" 2>/dev/null || true
    
    # Создание .icns файла
    if iconutil -c icns "$ICONSET_DIR" -o "$APP_BUNDLE/Contents/Resources/icon.icns" 2>/dev/null; then
        log_info "Иконка успешно создана"
    else
        log_warn "Не удалось создать иконку, продолжаю без нее"
    fi
    
    # Очистка временной директории
    rm -rf "$ICONSET_DIR"
}

# Создание Info.plist
create_info_plist() {
    log_info "Создание Info.plist для macOS $MACOS_MIN_VERSION+..."
    
    cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$BINARY_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>com.yoshapihoff.smart-clipboard</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSMinimumSystemVersion</key>
    <string>$MACOS_MIN_VERSION</string>
    <key>NSAppleEventsUsageDescription</key>
    <string>Smart Clipboard needs access to system events for clipboard management</string>
    <key>NSAppleScriptEnabled</key>
    <true/>
    <key>NSCameraUsageDescription</key>
    <string>This app does not require camera access</string>
    <key>NSMicrophoneUsageDescription</key>
    <string>This app does not require microphone access</string>
    <key>NSLocationWhenInUseUsageDescription</key>
    <string>This app does not require location access</string>
    <key>NSDesktopFolderUsageDescription</key>
    <string>Smart Clipboard needs desktop access for clipboard operations</string>
    <key>NSDocumentsFolderUsageDescription</key>
    <string>Smart Clipboard needs documents access for clipboard operations</string>
    <key>NSDownloadsFolderUsageDescription</key>
    <string>Smart Clipboard needs downloads access for clipboard operations</string>
    <key>NSSystemAdministrationUsageDescription</key>
    <string>Smart Clipboard requires system administration access for clipboard management</string>
    <key>CFBundleSupportedPlatforms</key>
    <array>
        <string>MacOSX</string>
    </array>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
    <key>NSRequiresAquaSystemAppearance</key>
    <false/>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
</dict>
</plist>
EOF
    
    log_info "Info.plist создан с поддержкой macOS $MACOS_MIN_VERSION+"
}

# Подписание приложения
sign_app() {
    log_info "Подписание приложения..."
    
    if codesign --force --deep --sign - "$APP_BUNDLE" 2>/dev/null; then
        log_info "Приложение успешно подписано"
    else
        log_warn "Не удалось подписать приложение, продолжаю без подписи"
    fi
}

# Основная функция
main() {
    log_info "Начало сборки $APP_NAME для macOS $MACOS_MIN_VERSION+..."
    log_info "Архитектура: $ARCH"
    
    check_dependencies
    build_binary
    create_app_structure
    create_icon
    create_info_plist
    sign_app
    
    # Вывод информации о собранном приложении
    log_info "Сборка завершена!"
    log_info "Приложение доступно по пути: $APP_BUNDLE"
    log_info "Архитектура: $ARCH"
    log_info "Минимальная версия macOS: $MACOS_MIN_VERSION"
    log_info "Вы можете запустить его двойным кликом или переместить в папку Applications"
    
    # Проверка бинарного файла
    if [ -f "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME" ]; then
        file "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME"
    fi
}

# Функция справки
show_help() {
    echo "Использование: $0 [arm64|x86_64|universal]"
    echo ""
    echo "Архитектуры:"
    echo "  arm64      - Apple Silicon (M1, M2, M3 и т.д.)"
    echo "  x86_64     - Intel Mac"
    echo "  universal  - Universal Binary (по умолчанию)"
    echo ""
    echo "Примеры:"
    echo "  $0              # Сборка Universal Binary"
    echo "  $0 arm64         # Сборка только для Apple Silicon"
    echo "  $0 x86_64        # Сборка только для Intel"
}

# Обработка аргументов
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Запуск основной функции
main "$@"
