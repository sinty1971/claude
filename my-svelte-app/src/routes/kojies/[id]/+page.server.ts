import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { createGrpcClient } from '$lib/grpc-client';
import {
  KojiService,
  GetKojiRequestSchema,
} from '../../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async ({ params }) => {
  const client = createGrpcClient(KojiService);

  try {
    const response = await client.getKoji(
      create(GetKojiRequestSchema, { targetId: params.id })
    );

    return {
      koji: response.koji ?? null,
    };
  } catch (err) {
    console.error('工事情報取得エラー:', err);
    throw error(500, '工事情報の取得に失敗しました');
  }
};
