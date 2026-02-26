# Troubleshooting Guide

## Authentication Issues

### "OAuth client credentials missing"

```bash
# Use default public client (release builds) or set a custom client:
export GDRV_CLIENT_ID="your-client-id"
export GDRV_CLIENT_SECRET="your-client-secret" # only if required by your client type
gdrv auth login
```

### "Browser not opening"

Use manual fallback or device code flow:
```bash
gdrv auth login --no-browser
gdrv auth device
```

### "Invalid credentials"

Re-authenticate:
```bash
gdrv auth logout
gdrv auth login
```

### "Missing required scope"

```bash
gdrv auth status
gdrv auth login --preset workspace-full
```

### "invalid_grant" errors

If your OAuth consent screen is in testing mode, refresh tokens expire after 7 days. Re-authenticate or move to production mode.

## Permission Errors

### "Insufficient permissions"

Check your OAuth scopes and Shared Drive access:
```bash
gdrv auth status
```

You may need to re-authenticate with a broader scope preset:
```bash
gdrv auth login --preset workspace-full
```

## Path Resolution

### "File not found"

Use file IDs for Shared Drives:
```bash
gdrv files list --drive-id <drive-id>
```

## Rate Limiting

The CLI automatically handles rate limits with exponential backoff. If you continue to see rate limit errors:

1. Reduce the frequency of API calls
2. Use pagination with smaller page sizes
3. Check Google's API quotas for your project

## Performance Issues

### Slow uploads/downloads

- Check your network connection
- Use resumable uploads for large files
- Consider using `--quiet` to reduce output overhead

### Large result sets

Use `--paginate` with `--limit` for controlled pagination:
```bash
# Get 50 items at a time
gdrv files list --limit 50 --json
# Then use nextPageToken from response
```

## Admin SDK Issues

### "Domain-wide delegation not enabled"

Admin SDK operations require service account authentication with domain-wide delegation. See [API-GUIDE.md](API-GUIDE.md#admin-sdk-operations) for setup instructions.

### "Not authorized to access this resource"

Ensure the service account has been authorized for the required scopes in the Google Workspace Admin Console.

## Getting Help

If you encounter issues not covered here:

1. Check the [API-GUIDE.md](API-GUIDE.md) for command-specific documentation
2. Review the [Authentication Guide](AUTHENTICATION.md) for auth-related issues
3. Run with `--help` to see all available flags for a command
4. Check exit codes for programmatic error handling
