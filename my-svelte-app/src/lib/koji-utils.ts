import { timestampDate } from "@bufbuild/protobuf/wkt";
import type { Koji } from "../gen/grpc/v1/toyotachikuro_pb";

/**
 * 工事のステータスを開始日と終了日から判定する
 * @param koji 工事オブジェクト
 * @returns ステータス文字列（"予定" | "進行中" | "完了" | "不明"）
 */
export function generateKojiStatus(koji?: Koji | null): string {
  if (!koji?.start || !koji?.mfEnd) {
    return "不明";
  }

  const startDate = timestampDate(koji.start);
  const endDate = timestampDate(koji.mfEnd);
  
  if (Number.isNaN(startDate.getTime()) || Number.isNaN(endDate.getTime())) {
    return "不明";
  }

  const now = new Date();

  if (now < startDate) {
    return "予定";
  } else if (now > endDate) {
    return "完了";
  } else {
    return "進行中";
  }
}

/**
 * ステータスに応じたCSSクラス名を返す
 * @param status ステータス文字列
 * @returns CSSクラス名
 */
export function getKojiStatusClass(status: string): string {
  switch (status) {
    case "進行中":
      return "status-active";
    case "完了":
      return "status-completed";
    case "予定":
      return "status-scheduled";
    default:
      return "status-unknown";
  }
}
