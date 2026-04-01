import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { resetSession } from '../store/slices/sessionSlice';
import { joinSession } from '../store/slices/sessionSlice';
import Logo from '../components/ui/Logo/Logo';
import Button from '../components/ui/Button/Button';
import styles from './JoinPage.module.css';

export default function JoinPage() {
  const { pin: pinParam } = useParams<{ pin?: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { isLoading, error, joinData } = useAppSelector((s) => s.session);

  
  const [pinDigits, setPinDigits] = useState<string[]>(
    Array.from({ length: 6 }, (_, i) => pinParam?.[i] ?? '')
  );
  const pin = pinDigits.join('').replace(/ /g, '');
  const [nickname, setNickname] = useState('');
  const [fieldErrors, setFieldErrors] = useState<{ pin?: string; nickname?: string }>({});
  const [resetDone, setResetDone] = useState(false);

  
  const pinRefs = useRef<(HTMLInputElement | null)[]>([]);

  
  
  useEffect(() => {
    dispatch(resetSession());
    
    try {
      Object.keys(localStorage)
        .filter((k) =>
          k.startsWith('sb_ws_signal_') ||
          k.startsWith('sb_answered_') ||
          k.startsWith('sb_current_question_')
        )
        .forEach((k) => localStorage.removeItem(k));
    } catch {  }
    sessionStorage.removeItem('sb_score');
    sessionStorage.removeItem('sb_correct');
    sessionStorage.removeItem('sb_total');
    setResetDone(true);
  }, [dispatch]);

  
  
  useEffect(() => {
    if (!resetDone || !joinData) return;
    sessionStorage.setItem('sb_nickname', nickname.trim());
    navigate(`/session/${joinData.sessionId}/waiting`, {
      state: {
        participantId: joinData.participantId,
        quizTitle: joinData.quiz.title,
      },
    });
  }, [joinData, navigate, nickname, resetDone]);

  const handlePinChange = (idx: number, value: string) => {
    if (!/^\d*$/.test(value)) return;
    const digit = value.slice(-1);
    const newDigits = [...pinDigits];
    newDigits[idx] = digit;
    setPinDigits(newDigits);
    setFieldErrors((p) => ({ ...p, pin: undefined }));
    if (digit && idx < 5) {
      pinRefs.current[idx + 1]?.focus();
    }
  };

  const handlePinKeyDown = (idx: number, e: React.KeyboardEvent) => {
    if (e.key === 'Backspace') {
      if (!pinDigits[idx] && idx > 0) {
        pinRefs.current[idx - 1]?.focus();
        const newDigits = [...pinDigits];
        newDigits[idx - 1] = '';
        setPinDigits(newDigits);
      } else {
        const newDigits = [...pinDigits];
        newDigits[idx] = '';
        setPinDigits(newDigits);
      }
    }
    if (e.key === 'ArrowLeft' && idx > 0) pinRefs.current[idx - 1]?.focus();
    if (e.key === 'ArrowRight' && idx < 5) pinRefs.current[idx + 1]?.focus();
  };

  const handlePinPaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const text = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const newDigits = Array.from({ length: 6 }, (_, i) => text[i] ?? '');
    setPinDigits(newDigits);
    if (text.length < 6) pinRefs.current[text.length]?.focus();
    else pinRefs.current[5]?.focus();
  };

  const validate = () => {
    const errs: typeof fieldErrors = {};
    const cleanPin = pin.replace(/\s/g, '');
    if (cleanPin.length < 6) errs.pin = 'Введите 6-значный PIN';
    if (!nickname.trim()) errs.nickname = 'Введите никнейм';
    else if (nickname.trim().length < 2) errs.nickname = 'Минимум 2 символа';
    setFieldErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    dispatch(joinSession({ pin: pin.trim(), nickname: nickname.trim() }));
  };

  return (
    <div className={styles.page}>
      <div className={styles.blob1} aria-hidden="true" />
      <div className={styles.blob2} aria-hidden="true" />

      <div className={styles.card}>
        <div className={styles.cardHeader}>
          <Logo size="lg" />
          <p className={styles.subtitle}>Введите данные, чтобы вступить в битву</p>
        </div>

        <form onSubmit={handleSubmit} className={styles.form} noValidate>
          <div className={styles.field}>
            <label className={styles.label}>PIN-код игры</label>
            <div className={styles.pinGroup}>
              <div className={styles.pinTriple}>
                {[0, 1, 2].map((idx) => (
                  <input
                    key={idx}
                    ref={(el) => { pinRefs.current[idx] = el; }}
                    className={[styles.pinCell, fieldErrors.pin ? styles.pinCellError : ''].join(' ')}
                    type="text"
                    inputMode="numeric"
                    maxLength={1}
                    value={pinDigits[idx] || ''}
                    onChange={(e) => handlePinChange(idx, e.target.value)}
                    onKeyDown={(e) => handlePinKeyDown(idx, e)}
                    onPaste={handlePinPaste}
                    onFocus={(e) => e.target.select()}
                    autoFocus={idx === 0}
                  />
                ))}
              </div>
              <span className={styles.pinDivider}>·</span>
              <div className={styles.pinTriple}>
                {[3, 4, 5].map((idx) => (
                  <input
                    key={idx}
                    ref={(el) => { pinRefs.current[idx] = el; }}
                    className={[styles.pinCell, fieldErrors.pin ? styles.pinCellError : ''].join(' ')}
                    type="text"
                    inputMode="numeric"
                    maxLength={1}
                    value={pinDigits[idx] || ''}
                    onChange={(e) => handlePinChange(idx, e.target.value)}
                    onKeyDown={(e) => handlePinKeyDown(idx, e)}
                    onPaste={handlePinPaste}
                    onFocus={(e) => e.target.select()}
                  />
                ))}
              </div>
            </div>
            {fieldErrors.pin && <span className={styles.error}>{fieldErrors.pin}</span>}
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Никнейм</label>
            <input
              className={[styles.nicknameInput, fieldErrors.nickname ? styles.nicknameInputError : ''].join(' ')}
              type="text"
              placeholder="Введите ваше имя"
              value={nickname}
              onChange={(e) => { setNickname(e.target.value); setFieldErrors((p) => ({ ...p, nickname: undefined })); }}
              maxLength={20}
            />
            {fieldErrors.nickname && <span className={styles.error}>{fieldErrors.nickname}</span>}
          </div>
          {error && <div className={styles.alertError}>{error}</div>}

          <Button type="submit" fullWidth size="lg" isLoading={isLoading}>
            Войти в комнату →
          </Button>
        </form>

        <p className={styles.footerBrand}>
          SMART <strong>BATTLE</strong>
        </p>
      </div>
    </div>
  );
}
