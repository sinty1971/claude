import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { createGrpcClient } from '$lib/grpc-client';
import {
  MemberService,
  GetMembersRequestSchema,
} from '../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async () => {
  const client = createGrpcClient(MemberService);

  try {
    const response = await client.getMembers(create(GetMembersRequestSchema, {}));

    const membersList = Object.values(response.members ?? {});

    // 名前でソート
    membersList.sort((a, b) => {
      const nameA = a.name || '名称未設定';
      const nameB = b.name || '名称未設定';
      return nameA.localeCompare(nameB, 'ja');
    });

    return {
      members: membersList,
    };
  } catch (err) {
    console.error('メンバー一覧取得エラー:', err);
    throw error(500, 'メンバー一覧の取得に失敗しました');
  }
};
