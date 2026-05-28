# Incident Response

> Operational playbook for security incidents in Ancient Coins. [SECURITY.md](../SECURITY.md) defines the reporting channel and disclosure window; this document defines what happens after a report or alert exists.

## Severity classification

| Severity | Meaning | Ancient Coins examples |
|---|---|---|
| P0 | Active compromise or immediately exploitable issue with user-impacting exposure. Treat as stop-the-line work. | Confirmed auth bypass, leaked JWT signing secret, publicly reachable API key dump, XSS with token theft, exposed user data from a public endpoint. |
| P1 | High-risk issue with credible exploitation path, but no confirmed active abuse yet. | Dependency advisory with a working exploit in a deployed path, mutable CI action compromise suspicion, broken WebAuthn origin validation in production, secrets committed to `main` but not yet abused. |
| P2 | Moderate-risk issue that should be remediated promptly but does not require immediate service interruption. | Missing cache-control headers on auth responses, non-root container hardening gap, missing body-size limits, low-confidence suspicious report needing validation. |
| P3 | Low-risk issue, hardening work, or accepted platform limitation. | Documentation-only correction, noisy false positive from scanners, browser-memory limitations around password clearing. |

## Roles and responsibilities

### Current solo-project model

- **Incident Commander:** Brian (`briandenicola`) owns triage, containment, remediation decisions, and external communication.
- **Primary operator:** Brian performs repo, hosting, secret rotation, and release actions.
- **Documentation owner:** update [threat-model.md](threat-model.md), [security-principles.md](security-principles.md), and the post-incident record before closing the incident.

### If collaborators are added later

- Add named backups for Incident Commander and release operator.
- Route notifications through CODEOWNERS / repository maintainers.
- Document who can approve hotfix merges, secret rotation, and advisory publication.

## Detection signals

| Signal | Source | What it indicates | First action |
|---|---|---|---|
| Secret scan failure | `gitleaks` in local pre-commit or CI | Credential-like content entered the diff or repo history surface. | Validate whether it is real, rotate if real, and quarantine the branch if needed. |
| Weekly security scan | `.github/workflows/security-scan.yml` weekly cron | Dependency or config drift surfaced outside normal feature work. | Open or update a finding in [threat-model.md](threat-model.md) and prioritize by severity. |
| Dependency dashboard / alerts | Dependabot dashboard, GitHub alerts, package ecosystem advisories | Known vulnerable package or action version. | Assess reachability, exposure, and available patched version. |
| Private report | [SECURITY.md](../SECURITY.md) / GitHub Security Advisories | External researcher or user found a security issue. | Acknowledge within the 72-hour SLA and begin triage. |
| User or operator report | Issues, email, direct observation, suspicious runtime behavior | Potential incident not yet tied to a scanner finding. | Reproduce, preserve evidence, and classify severity. |

## Response timeline and checklist

### 1. Acknowledge

- Confirm receipt within the [SECURITY.md](../SECURITY.md) SLA (72 hours for private reports).
- Assign a severity (P0–P3).
- Open a private working note or GitHub Security Advisory draft if disclosure may be required.
- Record the initial hypothesis, reporter, and affected version/branch.

### 2. Contain

- Revoke or rotate exposed secrets, tokens, or API keys.
- Disable affected workflows, endpoints, or features if leaving them live increases blast radius.
- Freeze unrelated deploys until the incident is understood.
- Preserve evidence: commit SHA, workflow run URL, logs, screenshots, reproduction steps.

### 3. Investigate

- Determine scope: which service, which users, which branches, which secrets, which time window.
- Reproduce safely in a non-public path when possible.
- Check [threat-model.md](threat-model.md) for a matching historical finding and update status if the risk became real.
- Decide whether the issue is a true incident, a near miss, or a false positive.

### 4. Remediate

- Ship the smallest safe fix first, then follow with hardening if needed.
- Add or update tests, scanners, or config so the same class of issue is harder to reintroduce.
- If the fix changes architecture or policy, open/update an ADR and [security-principles.md](security-principles.md).

### 5. Communicate

- For externally reportable issues, publish or prepare a GitHub Security Advisory.
- Tell affected users what happened, what data or capability was at risk, what was fixed, and what they need to do (for example rotate keys or re-login).
- Keep the public statement factual; do not speculate beyond verified scope.

### 6. Review and close

- Capture the post-incident write-up using the template below.
- Update [threat-model.md](threat-model.md) status and any new standing controls in [security-principles.md](security-principles.md).
- Confirm disclosure timing still complies with [SECURITY.md](../SECURITY.md)'s 30-day coordinated disclosure guidance.

## Communication template

When an issue needs a formal advisory, use GitHub Security Advisories and include at least:

- **Title:** short, concrete summary (`Stored XSS in AI analysis rendering`)
- **Severity:** P0/P1/P2/P3 plus CVSS if later needed
- **Affected surface:** API, frontend, agent, CI, dependency, or deployment path
- **Affected versions / branches:** `main`, released image tags, or commit range
- **Impact:** what an attacker could do
- **Conditions required:** auth required? admin only? public route?
- **Mitigation / fix:** patched commit, workaround, config change, rotation instructions
- **Credits:** reporter name or handle if they want attribution
- **Timeline:** reported, acknowledged, fixed, disclosed

## Post-incident review template

Store future write-ups under the pattern `docs/post-incident/YYYY-MM-DD-incident-slug.md`.

Recommended structure:

```md
# Incident: <short title>

- Date discovered:
- Severity:
- Report source:
- Affected versions:
- Incident commander:

## Summary
## Timeline
## Root cause
## Impact
## Containment
## Remediation
## Follow-up actions
## Links
```

## Related documents

- [SECURITY.md](../SECURITY.md)
- [threat-model.md](threat-model.md)
- [security-principles.md](security-principles.md)
- [references.md](references.md)
- [ADR process](adr/README.md)
