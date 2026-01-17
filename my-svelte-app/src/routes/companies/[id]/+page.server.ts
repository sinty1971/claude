import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { createGrpcClient } from '$lib/grpc-client';
import {
  CompanyService,
  GetCompanyRequestSchema,
  GetCompanyCategoriesRequestSchema,
} from '../../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async ({ params }) => {
  const client = createGrpcClient(CompanyService);

  try {
    const [companyResponse, categoryResponse] = await Promise.all([
      client.getCompany(create(GetCompanyRequestSchema, { targetId: params.id })),
      client.getCompanyCategories(create(GetCompanyCategoriesRequestSchema, {})),
    ]);

    return {
      company: companyResponse.company ?? null,
      categories: categoryResponse.categories ?? [],
    };
  } catch (err) {
    console.error('会社情報取得エラー:', err);
    throw error(500, '会社情報の取得に失敗しました');
  }
};
