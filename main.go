package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// .rcgen
type Config struct {
	PathToFolder string
}

func main() {
	f, err := os.Open(".rcconfig")

	if err != nil {
		log.Fatalf("could not find .rcgen in root")
	}

	buff := make([]byte, 4096)

	n, err := f.Read(buff)

	if err != nil {
		log.Fatal("failed to read .rcgen")
	}

	if n == 0 {
		log.Fatal("rcgen is empty")
	}

	config, err := parseBuffer(buff, n)

	if err != nil {
		log.Fatal(err)
	}

	level := flag.Uint64("L", 0, "specifies chosen component level")
	name := flag.String("N", "DefaultComponent", "specifies component filename (case sensitive)")

	flag.Parse()

	if err = buildFolder(int64(*level), *name, config.PathToFolder); err != nil {
		log.Fatal(err)
	}
}

func buildFolder(level int64, name string, root string) error {
	_root := strings.Trim(root, "/")
	folderBaseName := fmt.Sprintf("%v/L%v", _root, level)
	utilsFolder := fmt.Sprintf("%v/%v", _root, "utils")
	cnFilePath := fmt.Sprintf("%v/cn.ts", utilsFolder)
	folderComponentName := fmt.Sprintf("%v/%v", folderBaseName, name)
	rootIndexFile := fmt.Sprintf("%v/index.ts", folderBaseName)

	if _, err := os.Stat(utilsFolder); os.IsNotExist(err) {
		if err := os.Mkdir(utilsFolder, os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(cnFilePath); os.IsNotExist(err) {
		if err := createFile(utilsFolder, "cn", ".ts", cnFile, false); err != nil {
			return err
		} else {
			log.Printf("created file at %v\n", cnFilePath)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%v/index.ts", _root)); os.IsNotExist(err) {
		if err := createFile(_root, "index", ".ts", mainIndexFile, false); err != nil {
			return err
		} else {
			log.Printf("created file at %v/index.ts\n", _root)
		}
	}

	if _, err := os.Stat(folderBaseName); os.IsNotExist(err) {
		if err := os.Mkdir(folderBaseName, os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(folderComponentName); os.IsNotExist(err) {
		if err := os.Mkdir(folderComponentName, os.ModePerm); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%v already exists", folderComponentName)
	}

	if err := createFile(folderComponentName, name, ".tsx", reactComponent, false); err != nil {
		return err
	} else {
		log.Printf("created file at %v/%v.tsx\n", folderComponentName, name)
	}

	if err := createFile(folderComponentName, name, ".stories.tsx", storybookComponent, false); err != nil {
		return err
	} else {
		log.Printf("created file at %v/%v.stories.tsx\n", folderComponentName, name)
	}

	if err := createFile(folderComponentName, name, ".test.tsx", testingFile, false); err != nil {
		return err
	} else {
		log.Printf("created file at %v/%v.test.tsx\n", folderComponentName, name)
	}

	if err := createFile(folderComponentName, name, ".ts", folderIndexFile, true); err != nil {
		return err
	} else {
		log.Printf("created file at %v/index.ts\n", folderComponentName)
	}

	if _, err := os.Stat(rootIndexFile); os.IsNotExist(err) {
		f, err := os.Create(rootIndexFile)

		if err != nil {
			return err
		} else {
			n, err := f.WriteString("// register all your components here as named exports")

			if n == 0 {
				return fmt.Errorf("failed to write to %v", rootIndexFile)
			}

			if err != nil {
				return err
			}

			log.Printf("created file at %v\n", rootIndexFile)
		}
	}

	return nil
}

func parseBuffer(buff []byte, n int) (*Config, error) {
	_buff := buff[0:n]
	str := strings.Trim(string(_buff), " ")

	if strings.Contains(str, "\"") || strings.Contains(str, "'") {
		return nil, errors.New("root string cannot contain quotes")
	}

	if !strings.HasPrefix(str, "root") {
		return nil, errors.New("could not find root token")
	}

	tokens := strings.Split(str, "=")

	if len(tokens) != 2 {
		return nil, errors.New("could not find key-value pair for root")
	}

	root := tokens[1]

	return &Config{
		PathToFolder: root,
	}, nil
}

func createFile(folderComponentName string, componentName string, ext string, fileContentGenerator func(c string) string, isIndex bool) error {
	filename := componentName

	if isIndex {
		filename = "index"
	}

	f, err := os.Create(fmt.Sprintf("%v/%v.%v", folderComponentName, filename, strings.Trim(ext, ".")))

	if err != nil {
		log.Println(err)
	}

	n, err := f.WriteString(fileContentGenerator(componentName))

	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("failed to write contents to file")
	}

	return nil
}

func reactComponent(componentName string) string {
	lowerCaseComponentName := fmt.Sprintf("%v%v", strings.ToLower(componentName[0:1]), componentName[1:])

	lines := []string{
		"import cn from \"@/utils/cn\";\n",
		"import { VariantProps, cva } from \"class-variance-authority\";\n",
		"import { ButtonHTMLAttributes, forwardRef } from \"react\";\n\n",
		fmt.Sprintf("export const %vVariants = ", lowerCaseComponentName),
		"cva(\"px-3 py-1 border-2 border-transparent\", {\n\tvariants: {\n\t\tstate: {\n\t\t\tsolid: \"bg-emerald-300 text-slate-800\",\n\t\t\toutline: \"border-emerald-500 text-emerald-900\",\n\t\t\twarning: \"bg-red-200 text-red-800\",\n\t\t\tghost: \"text-slate-700 hover:bg-slate-100\",\n\t\t},\n\t\tsize: {\n\t\t\tcontent: \"\",\n\t\t\tstretch: \"w-full\",\n\t\t},\n\t},\n\tdefaultVariants: {\n\t\tstate: \"solid\",\n\t\tsize: \"content\",\n\t},\n});\n\n",

		fmt.Sprintf("export interface %vProps extends ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof %vVariants> {}\n\n", componentName, lowerCaseComponentName),
		fmt.Sprintf("const %v = forwardRef<HTMLButtonElement, %vProps>((props, ref) => {\n", componentName, componentName),
		"\tconst { state, size, className, ...rest } = props;\n",
		fmt.Sprintf("\tconst style = cn(%vVariants({ size, state, className }));\n\n", lowerCaseComponentName),
		"\treturn <button ref={ref} className={style} {...rest} />\n",
		"});\n\n",
		fmt.Sprintf("%v.displayName = \"%v\"\n\n;", componentName, componentName),

		fmt.Sprintf("export default %v;\n", componentName),
	}

	content := strings.Join(lines, "")
	return content
}

func storybookComponent(componentName string) string {
	lines := []string{
		"import { Meta, StoryObj } from \"@storybook/react\";\n",
		fmt.Sprintf("import %v from \"./%v\";\n\n", componentName, componentName),
		fmt.Sprintf("const meta: Meta<typeof %v> = {\n", componentName),
		fmt.Sprintf("\tcomponent: %v,\n", componentName),
		"\targs: {\n\t\tchildren: \"A demo button\",\n\t\tstate: \"solid\",\n\t\tsize: \"content\",\n\t},\n};\n\n", fmt.Sprintf("type Story = StoryObj<typeof %v>;\n\n", componentName), "export const Solid: Story = {\n\targs: {},\n};\n\nexport const Warning: Story = {\n\targs: {\n\t\tstate: \"warning\",\n\t},\n};\n\nexport const Outlined: Story = {\n\targs: {\n\t\tstate: \"outline\",\n\t},\n};\n\nexport const Ghost: Story = {\n\targs: {\n\t\tstate: \"ghost\",\n\t},\n};\n\nexport const FullWidth: Story = {\n\targs: {\n\t\tstate: \"solid\",\n\t\tsize: \"stretch\",\n\t},\n};\n\nexport default meta;",
	}

	content := strings.Join(lines, "")

	return content
}

func testingFile(componentName string) string {
	lines := []string{
		"import \"@testing-library/jest-dom\";\n",
		"import { render, screen } from \"@testing-library/react\";\n",
		"import userEvent from \"@testing-library/user-event\";\n",
		fmt.Sprintf("import %v from \"./%v\";\n\n", componentName, componentName),
		fmt.Sprintf("it(\"should render a div with text `%v`\", async () => {\n", componentName),
		"\t// ARRANGE\n",
		fmt.Sprintf("\trender(<%v />);\n\n", componentName),
		"\t// ACT\n",
		fmt.Sprintf("\tconst el = screen.getByText(\"%v\");\n", componentName),
		"\tawait userEvent.click(el);\n\n",
		"\t// ASSERT\n",
		"\texpect(el).toBeInTheDocument();\n});\n",
	}

	content := strings.Join(lines, "")

	return content
}

func folderIndexFile(componentName string) string {
	return fmt.Sprintf("import %v from \"./%v\";\n\nexport default %v;\n", componentName, componentName, componentName)
}

func mainIndexFile(componentName string) string {
	return "// export all your component levels here\n// for example: export * as L0 from \"./L0\";\n"
}

func cnFile(componentName string) string {
	return "import { ClassValue, clsx } from \"clsx\";\nimport { twMerge } from \"tailwind-merge\";\n\nexport default function cn(...inputs: ClassValue[]) {\n\treturn twMerge(clsx(inputs));\n}\n"
}
