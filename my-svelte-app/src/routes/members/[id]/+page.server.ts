import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { createGrpcClient } from '$lib/grpc-client';
import {
  MemberService,
  GetMemberRequestSchema,
} from '../../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async ({ params }) => {
  const client = createGrpcClient(MemberService);

  try {
    const response = await client.getMember(create(GetMemberRequestSchema, { targetId: params.id }));

    return {
      member: response.member ?? null,
    };
  } catch (err) {
    console.error('メンバー情報取得エラー:', err);
    throw error(500, 'メンバー情報の取得に失敗しました');
  }
};
