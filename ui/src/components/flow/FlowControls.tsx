import React from 'react';
import { useTranslation } from 'react-i18next';
import { Play, Square, FastForward, PlayCircle, PauseCircle, Save, RotateCcw, ToggleLeft, ToggleRight } from 'lucide-react';

interface FlowControlsProps {
  onRun: () => void;
  onRunTask: () => void;
  onStop: () => void;
  onResume: () => void;
  onStopAtTask: () => void;
  onSaveLayout: () => void;
  onResetLayout: () => void;
  onToggleAutoSaveLayout: () => void;
  autoSaveLayout: boolean;
  isFlowRunning: boolean;
  runPending: boolean;
  taskRunPending: boolean;
  stopPending: boolean;
  resumePending: boolean;
  stopAtTaskPending: boolean;
  canRunTask: boolean;
  canResume: boolean;
  canStopAtTask: boolean;
  stopAtTaskActive: boolean;
}

const FlowControls: React.FC<FlowControlsProps> = ({
  onRun,
  onRunTask,
  onStop,
  onResume,
  onStopAtTask,
  onSaveLayout,
  onResetLayout,
  onToggleAutoSaveLayout,
  autoSaveLayout,
  isFlowRunning,
  runPending,
  taskRunPending,
  stopPending,
  resumePending,
  stopAtTaskPending,
  canRunTask,
  canResume,
  canStopAtTask,
  stopAtTaskActive,
}) => {
  const { t } = useTranslation();

  return (
    <div className="flow-controls">
      <div className="flow-controls__group">
        <button
          className="flow-controls__button flow-controls__button--primary"
          onClick={onRun}
          disabled={runPending || isFlowRunning}
          title={t('buttons.run')}
        >
          <Play size={20} fill="currentColor" />
        </button>
        <button
          className="flow-controls__button flow-controls__button--stop-at"
          onClick={onStopAtTask}
          disabled={!canStopAtTask || stopAtTaskPending}
          data-active={stopAtTaskActive ? 'true' : 'false'}
          title={t('buttons.stopAtTask')}
        >
          <PauseCircle size={20} />
        </button>
        <button
          className="flow-controls__button"
          onClick={onStop}
          disabled={!isFlowRunning || stopPending}
          title={t('buttons.stopFlow')}
        >
          <Square size={20} fill="currentColor" />
        </button>
      </div>

      <div className="flow-controls__divider" />

      <div className="flow-controls__group">
        <button
          className="flow-controls__button"
          onClick={onRunTask}
          disabled={!canRunTask || taskRunPending || runPending || isFlowRunning}
          title={t('buttons.runTask')}
        >
          <PlayCircle size={20} />
        </button>
        <button
          className="flow-controls__button"
          onClick={onResume}
          disabled={!canResume || resumePending}
          title={t('buttons.resumeFromTask')}
        >
          <FastForward size={20} fill="currentColor" />
        </button>
      </div>

      <div className="flow-controls__divider" />

      <div className="flow-controls__group">
        <button
          className="flow-controls__button"
          onClick={onSaveLayout}
          title={t('buttons.saveLayout')}
        >
          <Save size={20} />
        </button>
        <button
          className="flow-controls__button flow-controls__button--reset"
          onClick={onResetLayout}
          title={t('buttons.resetLayout')}
        >
          <RotateCcw size={20} />
        </button>
        <button
          className="flow-controls__button flow-controls__button--toggle"
          onClick={onToggleAutoSaveLayout}
          data-active={autoSaveLayout ? 'true' : 'false'}
          title={autoSaveLayout ? t('buttons.autoSaveLayoutOn') : t('buttons.autoSaveLayoutOff')}
        >
          {autoSaveLayout ? <ToggleRight size={20} /> : <ToggleLeft size={20} />}
        </button>
      </div>
    </div>
  );
};

export default FlowControls;
