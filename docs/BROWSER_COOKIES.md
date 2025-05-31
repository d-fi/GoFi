# Browser Cookie Authentication

GoFi can automatically read the Deezer ARL cookie from your browser, making authentication easier.

## Supported Browsers

- **Chrome** (Windows, macOS, Linux)
- **Firefox** (Windows, macOS, Linux)
- **Microsoft Edge** (Windows, macOS, Linux)
- **Arc** (Windows, macOS, Linux)
- **Safari** (macOS only)

## Usage

To authenticate with Deezer using browser cookies:

```bash
./gofi auth deezer
```

This command will:
1. Search for the Deezer ARL cookie in all supported browsers
2. Validate the cookie
3. Save it to a `.env` file for future use
4. Set it in the current environment

## How It Works

The browser cookie reader:
- Accesses browser cookie databases (SQLite for Chrome/Firefox)
- Handles platform-specific encryption (Chrome encrypts cookies on some platforms)
- Safely reads cookies without modifying browser data
- Falls back to other browsers if one fails

## Security Notes

- Cookie data is read in read-only mode
- Temporary copies of cookie databases are used to avoid conflicts
- The ARL token is stored securely in your `.env` file
- Only the Deezer ARL cookie is accessed

## Manual Authentication

If automatic cookie reading fails, you can still authenticate manually:

1. Log into Deezer in your browser
2. Open Developer Tools (F12)
3. Go to Application/Storage → Cookies
4. Find the `arl` cookie for `deezer.com`
5. Set it as an environment variable:
   ```bash
   export DEEZER_ARL="your_arl_cookie_value"
   ```

## Troubleshooting

If cookie reading fails:
- Ensure you're logged into Deezer in at least one supported browser
- Check that the browser is closed (some browsers lock their cookie database while running)
- On macOS, you may need to grant terminal access to browser data
- Try different browsers if one doesn't work

## Platform-Specific Notes

### macOS
- Chrome/Edge cookies are encrypted using the macOS Keychain
- You may be prompted for keychain access
- Safari uses a binary cookie format (not fully supported yet)

### Windows
- Chrome/Edge cookies are encrypted using Windows DPAPI
- Full decryption support is planned for a future update

### Linux
- Chrome/Edge cookies use a simple encryption with a known key
- Firefox cookies are stored unencrypted