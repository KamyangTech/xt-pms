/**
 * 数字金额转中文大写（用于收付款凭证/发票等场景）
 * 支持最大到万亿（包含小数点后两位：角、分）
 *
 * @example
 *   toChineseUpperAmount(1234.56) // 壹仟贰佰叁拾肆元伍角陆分
 *   toChineseUpperAmount(1000)    // 壹仟元整
 *   toChineseUpperAmount(0)       // 零元整
 *   toChineseUpperAmount(10000)   // 壹万元整
 *   toChineseUpperAmount(0.07)    // 柒分
 *   toChineseUpperAmount(1000001) // 壹佰万零壹元整
 */
const DIGITS = ['零', '壹', '贰', '叁', '肆', '伍', '陆', '柒', '捌', '玖'];

/**
 * 把整数部分（>=1）转为中文大写，不含"元"字
 */
function convertInteger(n: number): string {
  const str = Math.floor(n).toString();
  const len = str.length;
  const parts: string[] = [];
  let lastWasZero = false;

  // 从最高位到最低位遍历
  for (let i = 0; i < len; i++) {
    const d = parseInt(str[i], 10);
    const pos = len - 1 - i; // 0=个, 1=拾, 2=佰, 3=仟, 4=万, 5=拾万, ...

    if (d === 0) {
      lastWasZero = true;
      continue;
    }

    if (lastWasZero && parts.length > 0) {
      const isSegmentStart = pos % 4 === 0;
      if (isSegmentStart) {
        // 跨段需要补零。例：1000001 = 100万 + 1，需要在 1 前面补"零"
        parts.push('零');
      } else {
        // 段内中间位需要补零。例：10010 = 1万 + 10，需要在 10 前面补"零"
        parts.push('零');
      }
    }

    const smallUnit = ['', '拾', '佰', '仟'][pos % 4] || '';
    let bigUnit = '';
    if (pos >= 4 && pos < 8) bigUnit = '万';
    else if (pos >= 8 && pos < 12) bigUnit = '亿';
    else if (pos >= 12 && pos < 16) bigUnit = '万亿';

    parts.push(DIGITS[d] + smallUnit + bigUnit);
    lastWasZero = false;
  }

  return parts.join('').replace(/零+$/g, '');
}

/**
 * 数字金额转中文大写
 * @param value 数字（最多两位小数）
 * @returns 中文大写字符串；空值返回空串
 */
export function toChineseUpperAmount(value: number | undefined | null): string {
  if (value === undefined || value === null || Number.isNaN(value)) return '';
  if (value < 0) return '';
  if (value === 0) return '零元整';

  // 截断到两位小数
  const num = Math.round(value * 100) / 100;
  const integerPart = Math.floor(num);
  const decimalPart = Math.round((num - integerPart) * 100);

  let result = '';
  if (integerPart > 0) {
    result = convertInteger(integerPart) + '元';
  }

  if (decimalPart === 0) {
    if (!result) return '零元整';
    result += '整';
    return result;
  }

  const jiao = Math.floor(decimalPart / 10);
  const fen = decimalPart % 10;

  if (!result) {
    // 整数部分为 0，仅有角分
    if (jiao === 0) {
      return `${DIGITS[fen]}分`;
    }
    if (fen === 0) {
      return `${DIGITS[jiao]}角整`;
    }
    return `${DIGITS[jiao]}角${DIGITS[fen]}分`;
  }

  if (jiao === 0) {
    return `${result}零${DIGITS[fen]}分`;
  }
  if (fen === 0) {
    return `${result}${DIGITS[jiao]}角整`;
  }
  return `${result}${DIGITS[jiao]}角${DIGITS[fen]}分`;
}
