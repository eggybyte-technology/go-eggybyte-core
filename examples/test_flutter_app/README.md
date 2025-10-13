# TestFlutterApp

Flutter application built with EggyByte standards.

## Getting Started

### Prerequisites

- Flutter SDK 3.16.0+
- Dart 3.2.0+
- Android Studio / Xcode (for mobile development)
- Chrome (for web development)

### Installation

1. Install dependencies:
   ```bash
   flutter pub get
   ```

2. Configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Run code generation:
   ```bash
   flutter pub run build_runner build
   ```

### Run

#### Mobile (Android/iOS)
```bash
flutter run
```

#### Web
```bash
flutter run -d chrome
```

### Build

#### Android APK
```bash
flutter build apk --release
```

#### iOS
```bash
flutter build ios --release
```

#### Web
```bash
flutter build web --release
```

## Project Structure

- `lib/` - Application source code
  - `config/` - Configuration files
  - `models/` - Data models
  - `services/` - API services
  - `screens/` - UI screens
  - `widgets/` - Reusable widgets
  - `providers/` - State management
- `assets/` - Images, fonts, and other assets
- `test/` - Unit and widget tests

## Development

### Code Generation

When you modify model files with JSON annotations:

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### Testing

Run all tests:

```bash
flutter test
```

Run tests with coverage:

```bash
flutter test --coverage
```

## API Integration

The app communicates with the EggyByte backend through RESTful APIs.

Base URL is configured in `lib/config/api_config.dart` and can be overridden via environment variables.

## License

Copyright © 2025 EggyByte Technology
