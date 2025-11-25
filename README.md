# GoPhish - Security Hardened Version

A security-enhanced fork of GoPhish with all known vulnerabilities patched.

---

## Official GoPhish Repository

This project is based on the official GoPhish open-source project:

- **Official Repository:** https://github.com/gophish/gophish
- **Official Website:** https://getgophish.com
- **Official Documentation:** https://docs.getgophish.com

---

## About GoPhish

GoPhish is an open-source phishing simulation platform designed to help organizations assess and improve their security awareness. It provides a comprehensive framework for launching simulated phishing campaigns, allowing security professionals to test how employees respond to phishing attempts in a controlled environment. The platform is widely used by penetration testers, security teams, and IT administrators to evaluate the human element of cybersecurity defenses.

The application offers a user-friendly web interface that simplifies the creation and management of phishing campaigns. Users can design custom email templates, create landing pages that mimic legitimate websites, and manage target groups of recipients. GoPhish tracks detailed metrics including email opens, link clicks, and credential submissions, providing valuable data for measuring the effectiveness of security awareness training programs.

GoPhish is built using the Go programming language, which provides excellent performance and cross-platform compatibility. The platform can be deployed on various operating systems including Windows, Linux, and macOS. It features a RESTful API that enables integration with other security tools and automation of campaign workflows. The self-hosted nature of GoPhish ensures that sensitive campaign data remains within the organization's infrastructure.

As a security testing tool, GoPhish plays a critical role in identifying vulnerabilities in an organization's human firewall. By conducting regular phishing simulations, organizations can identify employees who may need additional security training, measure improvements over time, and demonstrate compliance with security policies. The platform helps transform security awareness from a theoretical concept into measurable, actionable intelligence that strengthens overall organizational security posture.

---

## Quick Start

### Prerequisites

- Docker Desktop installed and running

### Installation

1. Clone this repository
2. Run the setup script:

**Windows:**
```batch
setup.bat
```

3. Access the application:
   - Admin Dashboard: https://localhost:8443

---

## Security Patch Status

All vulnerabilities identified during the security assessment have been patched.

### High Severity (5)

| # | Vulnerability | Status |
|---|---------------|--------|
| 1 | HTML Injection | PATCHED |
| 2 | Stored XSS | PATCHED |
| 3 | IDOR (Insecure Direct Object Reference) | PATCHED |
| 4 | Broken Access Control - Group Creation | PATCHED |
| 5 | Session Does Not Expire After Password Change | PATCHED |

### Medium Severity (5)

| # | Vulnerability | Status |
|---|---------------|--------|
| 6 | Lack of Security Headers | PATCHED |
| 7 | Session Does Not Expire After Logout | PATCHED |
| 8 | Vulnerable Component (Highcharts) | PATCHED |
| 9 | Vulnerable Component (Moment.js) | PATCHED |
| 10 | Vulnerable Component (ua-parser-js) | PATCHED |

### Low Severity (4)

| # | Vulnerability | Status |
|---|---------------|--------|
| 11 | Old CSRF Token Reuse | PATCHED |
| 12 | SameSite Flag Not Set | PATCHED |
| 13 | Weak Password Policy | PATCHED |
| 14 | Cleartext Credentials in Mail Server Config (CVE-2024-55196) | PATCHED |

**Total: 14 Vulnerabilities - All Patched**

---

## Acknowledgments

This project is a security-hardened fork of the original GoPhish project created by Jordan Wright. All credit for the core functionality goes to the original GoPhish team and contributors.

- Original Author: Jordan Wright
- Original Project: https://github.com/gophish/gophish
- License: MIT License

