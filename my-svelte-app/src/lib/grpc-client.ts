import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { GenService, GenServiceMethods } from '@bufbuild/protobuf/codegenv2';

/**
 * gRPC Connect のベース URL を取得
 * サーバーサイドでは process.env、クライアントサイドでは import.meta.env を使用
 */
function getBaseUrl(): string {
  if (typeof process !== 'undefined' && process.env?.VITE_CONNECT_BASE_URL) {
    return process.env.VITE_CONNECT_BASE_URL;
  }
  if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_CONNECT_BASE_URL) {
    return import.meta.env.VITE_CONNECT_BASE_URL;
  }
  return 'http://localhost:9090';
}

/**
 * gRPC Connect トランスポートを作成
 */
export function createGrpcTransport() {
  return createConnectTransport({
    baseUrl: getBaseUrl(),
    useBinaryFormat: true,
  });
}

/**
 * gRPC サービスクライアントを作成
 *
 * @example
 * ```ts
 * import { CompanyService } from '../gen/grpc/v1/toyotachikuro_pb';
 * const client = createGrpcClient(CompanyService);
 * ```
 */
export function createGrpcClient<T extends GenServiceMethods>(
  service: GenService<T>
) {
  const transport = createGrpcTransport();
  return createClient(service, transport);
}
