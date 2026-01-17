import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { createGrpcClient } from '$lib/grpc-client';
import { KojiService, GetKojiesRequestSchema } from '../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async () => {
  const client = createGrpcClient(KojiService);

  try {
    const response = await client.getKojies(create(GetKojiesRequestSchema, {}));
    const kojiesList = Object.values(response.kojies ?? {});

    // ソート処理（開始日の新しい順）
    kojiesList.sort((a, b) => {
      const aTime = a.start ? timestampDate(a.start).getTime() : -Infinity;
      const bTime = b.start ? timestampDate(b.start).getTime() : -Infinity;
      if (aTime !== bTime) return bTime - aTime;
      return (a.companyName || '').localeCompare(b.companyName || '', 'ja');
    });

    return {
      kojies: kojiesList,
    };
  } catch (err) {
    console.error('工事一覧取得エラー:', err);
    throw error(500, '工事一覧の取得に失敗しました');
  }
};
