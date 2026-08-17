const EMAIL_LOCAL_PATTERN = /^[A-Z0-9!#$%&'*+/=?^_{|}~.-]+$/i;
const DOMAIN_LABEL_PATTERN = /^[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?$/i;
const TOP_LEVEL_DOMAIN_PATTERN = /^[A-Z]{2,63}$/i;

export function isValidEmail(value: string) {
  const email = value.trim();
  if (!email || email.length > 254) return false;

  const at = email.indexOf("@");
  if (at <= 0 || at !== email.lastIndexOf("@")) return false;

  const local = email.slice(0, at);
  const domain = email.slice(at + 1);
  if (local.length > 64 || local.startsWith(".") || local.endsWith(".") || local.includes("..")) return false;
  if (!EMAIL_LOCAL_PATTERN.test(local) || domain.length > 253 || domain.includes("..")) return false;

  const labels = domain.split(".");
  if (labels.length < 2 || labels.some((label) => !DOMAIN_LABEL_PATTERN.test(label))) return false;
  return TOP_LEVEL_DOMAIN_PATTERN.test(labels.at(-1) ?? "");
}

export function isValidPhone(value: string) {
  const phone = value.trim();
  if (!phone || phone.length > 32 || !/^\+?[0-9().\s-]+$/.test(phone)) return false;
  const body = phone.startsWith("+") ? phone.slice(1) : phone;
  if (!/^(?:\d|\()/.test(body) || !/\d$/.test(body)) return false;
  if (!hasBalancedParentheses(phone)) return false;

  const international = phone.startsWith("+");
  const digits = phone.replace(/\D/g, "");
  if (international) {
    if (!/^[1-9]\d{7,14}$/.test(digits)) return false;
    if (digits.startsWith("1")) return digits.length === 11 && isValidNanp(digits.slice(1));
    return true;
  }

  const nationalDigits = digits.length === 11 && digits.startsWith("1") ? digits.slice(1) : digits;
  return isValidNanp(nationalDigits);
}

function isValidNanp(digits: string) {
  return /^[2-9]\d{2}[2-9]\d{6}$/.test(digits);
}

function hasBalancedParentheses(value: string) {
  let depth = 0;
  for (const character of value) {
    if (character === "(") depth++;
    if (character === ")") depth--;
    if (depth < 0 || depth > 1) return false;
  }
  return depth === 0;
}
