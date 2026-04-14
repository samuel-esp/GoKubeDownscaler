import Link from "@docusaurus/Link";

type ButtonProps = {
  name: string;
  to: string;
  className?: string;
  primary?: boolean;
};

export function Button({ name, to, className, primary = false }: ButtonProps) {
  const baseClasses =
    "rounded-md cursor-pointer text-xl font-bold py-2 px-8 text-center duration-200 transition-colors select-none whitespace-nowrap no-underline hover:no-underline block w-full border border-solid";

  const variantClasses = primary
    ? "bg-magenta hover:bg-magenta-hover active:bg-magenta-active border-magenta hover:border-magenta-hover active:border-magenta-active text-white hover:text-white dark:bg-gray-100 dark:hover:bg-gray-200 dark:active:bg-gray-300 dark:border-gray-100 dark:hover:border-gray-200 dark:text-slate-900 dark:hover:text-slate-900"
    : "bg-gray-100 hover:bg-gray-200 active:bg-gray-300 border-gray-100 hover:border-gray-200 active:border-gray-300 text-slate-900 hover:text-slate-900";

  return (
    <div className={className}>
      <Link
        draggable={false}
        className={`${baseClasses} ${variantClasses}`}
        to={to}
      >
        {name}
      </Link>
    </div>
  );
}
