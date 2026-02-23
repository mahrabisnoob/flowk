import { FlowLayoutSnapshot, fetchFlowLayout, saveFlowLayout, deleteFlowLayout } from '../api/client';

export type LayoutSnapshot = FlowLayoutSnapshot;
export type { FlowLayoutSnapshot };

export const fetchLayoutSnapshot = async (
  flowId: string,
  sourceName?: string
): Promise<FlowLayoutSnapshot | null> => {
  if (!flowId?.trim()) {
    return null;
  }
  try {
    return await fetchFlowLayout(flowId, sourceName);
  } catch (error) {
    console.warn('No se pudo cargar el layout del flow.', error);
    return null;
  }
};

export const saveLayoutSnapshot = async (
  flowId: string,
  sourceName: string | undefined,
  snapshot: FlowLayoutSnapshot
): Promise<void> => {
  if (!flowId?.trim()) {
    return;
  }
  try {
    await saveFlowLayout(flowId, sourceName, snapshot);
  } catch (error) {
    console.warn('No se pudo guardar el layout del flow.', error);
  }
};

export const deleteLayoutSnapshot = async (flowId: string, sourceName?: string): Promise<void> => {
  if (!flowId?.trim()) {
    return;
  }
  try {
    await deleteFlowLayout(flowId, sourceName);
  } catch (error) {
    console.warn('No se pudo eliminar el layout del flow.', error);
  }
};
