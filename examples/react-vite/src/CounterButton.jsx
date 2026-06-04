// This stays a normal JSX component to prove that Barkx can import and render existing React components.
export default function CounterButton({
  className = "",
  count,
  onIncrement,
  label = "React JSX +1",
  detail = `Global count: ${count}`,
}) {
  const classes = ["counter-button", className].filter(Boolean).join(" ");

  return (
    <button className={classes} onClick={onIncrement} type="button">
      <span className="button-label">{label}</span>
      <span className="button-detail">{detail}</span>
    </button>
  );
}
