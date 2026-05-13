/*
 * @autor: Gustavo Melo
 * @date: 2026-05-13
 * @description: Priemira aplicação em C# para aprender a sintaxe.
     O Objetivo é criar um objeto Pessoa com as propriedades Nome, Idade e Cpf, e data de nascimento,
     e criar uma classe que herda de Pessoa, chamada Aluno, com as propriedades Grade, e EscolaMatriculada.
     O objeto Aluno deve ter um método chamado MatricularEm, que recebe uma String como parâmetro e adiciona
     a propriedade EscolaMatriculada, se a propriedade estiver vazia. O método deve retornar void.
     O objeto Aluno deve ter um método chaamdo irParaEscola, que recebe uma String como parâmetro, e deve
     comparar com o atributo EscolaMatriculada se for diferente deve informar que está indo para a escola errada, senão, deve imprimir "indo para " + nome da EscolaMatriculada
     O objeto Aluno deve ter um método chamado ContarAteUmNumero, que recebe um número inteiro como parâmetro,
     e roda um loop for de 0 até o valor do parâmetro, somando os valores entre o intervalo informado. o retorno é uma lista de inteiros.
    O objeto pessoa deve ter um método chamado Apresentar se o atributo nome estiver preenchido deve retornar uma apresentação da instância.
    O objeto pessoa deve ter um método chamado Conversar que recebe uma String como parâmetro, e outra pessoa como parâmetro e imprime uma saudação, ex: "Gustavo disse: Oi Fulano | Fulano disse: Oi Gustavo"

    Ao final da criação dos objetos, na Main(), um aluno conversa com uma pessoa no meio do caminho e conta até 10, e até 100 de 2 em 2.

*/


public class Pessoa
{
    // Propriedades Autointituladas 
    private string Nome { get; set; } = string.Empty;
    private string Cpf { get; set; } = string.Empty;
    private DateTime DataDeNascimento { get; set; }
    private int Idade { get; set; } = 0;

    // Métodos Setters
    public void SetNome(string nome)
    {
        this.Nome = nome;
    }

    public void SetCpf(string cpf)
    {
        this.Cpf = cpf;
    }

    public void SetDataDeNascimento(DateTime dataDeNascimento)
    {
        this.DataDeNascimento = dataDeNascimento;
        this.Idade = DateTime.Now.Year - DataDeNascimento.Year;
    }

    // Métodos Getters
    public string GetNome()
    {
        return this.Nome;
    }

    public int GetIdade()
    {
        return this.Idade;
    }

    public DateTime GetDataDeNascimento()
    {
        return this.DataDeNascimento;
    }

    public string GetCpf()
    {
        return this.Cpf;
    }

    public void Apresentar()
    {
        Console.WriteLine("Olá, meu nome é " + Nome + " e eu tenho " + Idade + " anos (nasci em " + DataDeNascimento + ").");
    }

    public void Conversar(string mensagem, Pessoa pessoa)
    {
        Console.WriteLine(this.Nome + " disse: " + mensagem + " | " + pessoa.Nome + " disse: Oi " + this.Nome);
    }
}


public class Aluno : Pessoa
{
    private string Grade { get; set; } = string.Empty;
    private string EscolaMatriculada { get; set; } = string.Empty;


    // Métodos Setters
    public void SetGrade(string grade)
    {
        this.Grade = grade;
    }

    public void SetEscolaMatriculada(string escolaMatriculada)
    {
        this.EscolaMatriculada = escolaMatriculada;
    }

    // Métodos Getters
    public string GetGrade()
    {
        return this.Grade;
    }

    public string GetEscolaMatriculada()
    {
        return this.EscolaMatriculada;
    }

    public void MatricularEm(string escola)
    {
        this.EscolaMatriculada = escola;
    }

    public void irParaEscola(string escola)
    {
        if (this.EscolaMatriculada != escola)
        {
            Console.WriteLine(this.GetNome() + " parou, por que estava indo para a escola errada!");
        }
        else
        {
            Console.WriteLine(this.GetNome() + " está indo para " + this.EscolaMatriculada + "!");
        }
    }

    public List<int> ContarAteUmNumero(int numero, int intervalo = 1)
    {
        Console.WriteLine($"{this.GetNome()} está Contando de 0 até {numero} --> pulando de {intervalo} em {intervalo}!");
        List<int> numeros = new List<int>();
        for (int i = 0; i <= numero; i += intervalo)
        {
            numeros.Add(i);
        }
        // String.Join concatena todos os itens de um Array ou List em uma String
        Console.WriteLine(string.Join(", ", numeros));
        return numeros;
    }
}


namespace MeuPrimeiroAppCSharp
{
    public class Application
    {
        public static void Main(string[] args)
        {

            Console.WriteLine("Olá mundo donet!");
            Console.WriteLine("Para meu primeiro programa, resolvi criar um história em POO");


            // Criação de objetos é parecido com Java:            
            Pessoa psCarlos = new Pessoa();
            psCarlos.SetNome("Carlos");
            psCarlos.SetDataDeNascimento(new DateTime(1985, 3, 6)); // 06/03/1985
            psCarlos.SetCpf("12345678901");
            
            Pessoa psAntonia = new Pessoa();
            psAntonia.SetNome("Antonia");
            psAntonia.SetDataDeNascimento(new DateTime(1987, 2, 15)); // 15/02/1987
            psAntonia.SetCpf("12345678901");

            Aluno alPedro = new Aluno();
            alPedro.SetNome("Pedro");
            alPedro.SetDataDeNascimento(new DateTime(2008, 9, 18)); // 18/09/2008
            alPedro.SetCpf("12345678901");
            alPedro.SetGrade("Primeiro ano do Ensino Médio");

            Console.WriteLine("Carlos, Antonia e Pedro são amigos e se encontram em uma rua...");

            psCarlos.Apresentar(); // Apresenta o Carlos
            psAntonia.Apresentar(); // Apresenta a Antonia
            psCarlos.Conversar("Oi", psAntonia); // Conversa com a Antonia
            psAntonia.Conversar("Oi", psCarlos); // Conversa com o Carlos

            Console.WriteLine("Pedro se aproxima dos amigos, conversando com eles...");

            alPedro.Apresentar(); // Apresenta o Pedro
            alPedro.Conversar("Oi", psCarlos); // Conversa com o Carlos
            alPedro.Conversar("Oi", psAntonia); // Conversa com a Antonia

            Console.WriteLine($"Pedro fala que está indo se matricular na escola, para a {alPedro.GetGrade()}!");
            alPedro.SetEscolaMatriculada("CE Urbano Rocha"); // matricula no Urbano Rocha

            Console.WriteLine("Pedro resolveu ir para a escola, mas esqueceu para qual direção ficava o colégio...");
            alPedro.irParaEscola("CE Graça Aranha"); // Tenta ir para o Graça Aranha
            alPedro.irParaEscola("CE Urbano Rocha"); // Vai para o Urbano Rocha

            Console.WriteLine("Pedro decide contar até 10 enquanto caminha...");
            var numeros = alPedro.ContarAteUmNumero(10); // Conta até 10 pulando de 1 em 1

            Console.WriteLine("Depois decide contar até 100, pulando de 2 em 2...");
            var numerosPares = alPedro.ContarAteUmNumero(100, 2); // Conta até 100 pulando de 2 em 2
        
        }
    }
}

/*
* Considerações sobre o código:
* C# é muito parecido mesmo com Java na Sintaxe para aplicações básicas. Disso eu já sabia, mas agora pude
* constatar que é verdade. O código ficou bem limpo e organizado, e eu consegui entender tudo facilmente.
*
* Uma coisa me chamou atenção na criação dos objetos
*
*/