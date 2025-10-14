# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in EggyByte Core, please follow these steps:

1. **Do not** open a public GitHub issue
2. Email security details to: security@eggybyte.com
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact assessment
   - Suggested fix (if available)

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Within 30 days (depending on complexity)

## Security Best Practices

When using EggyByte Core in production:

1. **Keep Dependencies Updated**: Regularly update to the latest version
2. **Use HTTPS**: Always use HTTPS for production deployments
3. **Validate Input**: Validate all user inputs and external data
4. **Monitor Logs**: Enable structured logging and monitor for suspicious activity
5. **Network Security**: Use proper network segmentation and firewall rules
6. **Secrets Management**: Store sensitive configuration in secure secret management systems

## Security Features

EggyByte Core includes several built-in security features:

- **Structured Logging**: Redacts sensitive information from logs
- **Health Checks**: Kubernetes-compatible health probes
- **Graceful Shutdown**: Proper cleanup of resources
- **Context Propagation**: Request tracing and cancellation
- **Input Validation**: Built-in validation helpers

## Disclosure Policy

- Security vulnerabilities are disclosed privately first
- Public disclosure occurs after a fix is available
- Credit is given to security researchers who responsibly disclose vulnerabilities
- No legal action will be taken against security researchers who follow responsible disclosure practices
