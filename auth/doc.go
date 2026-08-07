//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package auth provides a client for the Meta system user and access token API.
//
// # Lifecycle
//
// System users follow a least-privilege pattern enforced by Meta:
//
//	REAL ADMIN (human)
//	  │
//	  ├─ 1. Create ADMIN SYSTEM USER ─── one per business, kept safe
//	  │     │
//	  │     ├─ 2. Install app on it ──── TOS acceptance (prerequisite for tokens)
//	  │     ├─ 3. Generate its token ─── now you can automate without the human token
//	  │     │
//	  │     └─ 4. Create REGULAR SYSTEM USER ─── one per access type, scoped
//	  │           │
//	  │           ├─ 5. Install app on it ──── same TOS step
//	  │           ├─ 6. Grant asset permissions ─── scoped to what it needs
//	  │           └─ 7. Generate its token ─── use this for daily API calls
//	  │
//	  └─ 8. Invalidate tokens ─── security escape hatch (can't delete users)
//
// Use the admin system user token only to manage other system users.
// Use regular system user tokens for all API calls. This way, if a token
// is compromised, the blast radius is limited to its scope.
//
// # Limits
//
// Standard access: 1 admin system user + 1 regular system user.
// Advanced access: 1 admin system user + 10 regular system users.
// The admin system user limit is always 1 — this is enforced by Meta to
// encourage least-privilege usage.
//
// # Token Lifecycle
//
// Tokens expire. Use [RefreshAccessToken] to extend without creating a new
// user, [RevokeAccessToken] to kill a single token, or
// [InvalidateSystemUserTokens] to kill all tokens for a user.
// [RotateAccessToken] provides atomic rotation: refresh → store new → revoke
// old, with no downtime.
//
// # Token Types Required
//
// Each endpoint requires a specific token type:
//
//	Endpoint                    Token Required
//	InstallApp                  Admin user or admin system user
//	GenerateAccessToken         Admin user or admin system user
//	CreateSystemUser            Admin user or admin system user
//	ListSystemUsers             Admin user or admin system user
//	UpdateSystemUser            Admin user or admin system user
//	InvalidateSystemUserTokens  Admin user or admin system user
//	RevokeAccessToken           Any valid token for the same app
//	RefreshAccessToken          The token being refreshed (fb_exchange_token)
//	AssignAdAccountPermissions  Admin user or admin system user
//	AssignPagePermissions       Admin user or admin system user
//	RetrieveSystemUserPerms     Admin user or admin system user
//	ClaimAppForBusiness         Admin user
//
// # Security Considerations
//
// [RevokeAccessToken] and [RefreshAccessToken] send the client_secret as a
// query parameter over GET. While this matches Meta's API design, secrets in
// URLs may be logged by intermediate proxies, CDNs, and load balancers. Always
// use TLS and consider the security posture of your network infrastructure.
package auth
