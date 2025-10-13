/// API Configuration for backend communication
class APIConfig {
  // Base URL for API endpoints
  static const String baseURL = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  // API endpoints
  static const String authEndpoint = '/v1/auth';
  static const String userEndpoint = '/v1/user';

  // Timeout configuration
  static const Duration timeout = Duration(seconds: 30);

  // Headers
  static Map<String, String> get defaultHeaders => {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      };
}
