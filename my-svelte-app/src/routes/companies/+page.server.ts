import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { create } from '@bufbuild/protobuf';
import { createGrpcClient } from '$lib/grpc-client';
import {
  CompanyService,
  CompanyCategoryService,
  GetCompaniesRequestSchema,
  GetCompanyCategoriesRequestSchema,
} from '../../gen/grpc/v1/toyotachikuro_pb';

export const load: PageServerLoad = async () => {
  const companyClient = createGrpcClient(CompanyService);
  const categoryClient = createGrpcClient(CompanyCategoryService);

  try {
    const [companiesResponse, categoriesResponse] = await Promise.all([
      companyClient.getCompanies(create(GetCompaniesRequestSchema, { forceReload: false })),
      categoryClient.getCompanyCategories(create(GetCompanyCategoriesRequestSchema, {})),
    ]);

    const companiesList = Object.values(companiesResponse.companies ?? {});
    
    // カテゴリマップを作成
    const categoryMap = new Map<number, string>();
    for (const category of categoriesResponse.categories ?? []) {
      categoryMap.set(category.index, category.name || '業種未設定');
    }

    // ソート処理
    companiesList.sort((a, b) => {
      const categoryCompare = (a.categoryIndex ?? 0) - (b.categoryIndex ?? 0);
      if (categoryCompare !== 0) return categoryCompare;
      const nameA = a.name || a.mfLongName || '名称未設定';
      const nameB = b.name || b.mfLongName || '名称未設定';
      return nameA.localeCompare(nameB, 'ja');
    });

    return {
      companies: companiesList,
      categoryMap: Object.fromEntries(categoryMap),
    };
  } catch (err) {
    console.error('会社一覧取得エラー:', err);
    throw error(500, '会社一覧の取得に失敗しました');
  }
};
