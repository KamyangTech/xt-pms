import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

// 配置 dayjs
dayjs.extend(utc);
dayjs.extend(timezone);

/**
 * 获取表头数据
 *
 * @export
 * @param {string[]} dateTime
 * @param {number} divideNum
 * @returns {string[]} timeArray
 */
export function getDateArray(dateTime: string[] = [], divideNum = 10): string[] {
  const timeArray: string[] = [];
  if (dateTime.length > 0) {
    for (let i = 0; i < divideNum; i++) {
      const dateAbsTime: number = (dayjs(dateTime[1]).tz('Asia/Shanghai').valueOf() - dayjs(dateTime[0]).tz('Asia/Shanghai').valueOf()) / divideNum;
      const enhandTime: number = dayjs(dateTime[0]).tz('Asia/Shanghai').valueOf() + dateAbsTime * i;
      timeArray.push(dayjs(enhandTime).tz('Asia/Shanghai').format('YYYY-MM-DD'));
    }
  }

  return timeArray;
}

/**
 * 获取随机数
 *
 * @param {number} [num]
 * @returns {number} resultNum
 */
export function getRandomArray(num = 100): number {
  let resultNum = Number((Math.random() * num).toFixed(0));

  if (resultNum <= 1) {
    resultNum = 1;
  }

  return resultNum;
}
