// This stays a normal JSX component to prove that Barkx can import and render existing React components.
export default function CounterButton({ className = "", count, onIncrement }) {
  const classes = ["counter-button", className].filter(Boolean).join(" ");

  return (
    <button className={classes} onClick={onIncrement} type="button">
      Clicked {count} times
    </button>
  );
}
