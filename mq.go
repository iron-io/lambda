package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
)

// type Cmd interface {
// 	Flags(...string) error // parse subcommand specific flags
// 	Args() error           // validate arguments
// 	Config() error         // configure env variables
// 	Usage() func()         // custom command help TODO(reed): all local now?
// 	Run()                  // cmd specific
// }

// It'd be better to abstract this out into two files, worker_command and mq_commands

type mqCmd struct {
	settings  config.Settings
	flags     *MqFlags
	token     *string
	projectID *string
}

func (mc *mqCmd) Config() error {
	mc.settings = config.Config("iron_mq")

	if *projectIDFlag != "" {
		mc.settings.ProjectId = *projectIDFlag
	}

	if *tokenFlag != "" {
		mc.settings.Token = *tokenFlag
	}

	if mc.settings.ProjectId == "" {
		return errors.New("did not find project id in any config files or env variables")
	}
	if mc.settings.Token == "" {
		return errors.New("did not find token in any config files or env variables")
	}

	if !isPipedOut() {
		fmt.Printf("%sConfiguring client\n", LINES)
	}
	// pName, err := mqProjectname(mc.settings)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf(`%s Project '%s' with id='%s'`, BLANKS, pName, mc.settings.ProjectId)
	// fmt.Println()

	return nil
}

type RmCmd struct {
	mqCmd

	queue_name string
}

type ListCmd struct {
	mqCmd

	//flags
	page    *int
	perPage *int
}

type ClearCmd struct {
	mqCmd

	queue_name string
}

type PeekCmd struct {
	mqCmd

	n          *int
	queue_name string
}

type PushCmd struct {
	mqCmd
	filename   *string
	messages   []string
	queue_name string
}

type PopCmd struct {
	mqCmd

	queue_name string
	n          *int
	outputfile *string
	file       *os.File
}

type ReserveCmd struct {
	mqCmd
	queue_name string
	n          *int
	timeout    *int
	outputfile *string
	file       *os.File
}

func printMessages(msgs []mq.Message) {
	for _, msg := range msgs {
		fmt.Printf("%s %q\n", msg.Id, msg.Body)
	}
}

// This is based on the format of func printMessages([]*mq.Message)
func readIds() ([]string, error) {
	var ids []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if len(message) > 19 {
			id := message[:19] // We want the first 19 characters of the line, since an id is 19 characters long
			ids = append(ids, id)
		}
	}
	return ids, scanner.Err()
}

// Check if stdout is being piped
func isPipedOut() bool {
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

func isPipedIn() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

func (l *ListCmd) Flags(args ...string) error {
	l.flags = NewMqFlagSet(l.Usage())

	l.page = l.flags.page()
	l.perPage = l.flags.perPage()

	err := l.flags.Parse(args)
	if err != nil {
		return err
	}
	return l.flags.validateAllFlags()
}

func (l *ListCmd) Args() error {
	return nil
}

func (l *ListCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq list [--perPage perPpage] [--page page]
    -perPage perPage: Amount of queues showed per page
    -page page: starting page number`)
		return
	}
}

func (l *ListCmd) Run() {
	queues, err := mq.List()
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	if isPipedOut() {
		for _, q := range queues {
			fmt.Println(q.Name)
		}
	} else {
		fmt.Println(LINES, "Listing queues")
		for _, q := range queues {
			fmt.Println(BLANKS, "*", q.Name)
		}
		// TODO: This can probably be put in its own function
		if tag, err := getHudTag(l.settings); err == nil {
			fmt.Printf("%s Go to hud-e.iron.io/mq/%s/projects/%s/queues for more info",
				BLANKS,
				tag,
				l.settings.ProjectId)
		}
		fmt.Println()
	}
}

type CreateCmd struct {
	mqCmd

	queue_name string
}

func (c *CreateCmd) Flags(args ...string) error {
	c.flags = NewMqFlagSet(c.Usage())
	err := c.flags.Parse(args)
	if err != nil {
		return err
	}

	return c.flags.validateAllFlags()
}

func (c *CreateCmd) Args() error {
	if c.flags.NArg() < 1 {
		return errors.New("create requires at least one argument\nusage: iron mq create QUEUE_NAME")
	}
	c.queue_name = c.flags.Arg(0)
	return nil
}

func (c *CreateCmd) Usage() func() {
	return func() {
		fmt.Println(`usage: iron mq create QUEUE_NAME`)
	}
}

func (c *CreateCmd) Run() {
	fmt.Printf("%sCreating queue \"%s\"\n", LINES, c.queue_name)
	q := mq.ConfigNew(c.queue_name, &c.settings)
	_, err := q.PushStrings("")
	if err != nil {
		fmt.Fprintln(os.Stderr, red(BLANKS, "create error: ", err))
		return
	}
	err = q.Clear()
	if err != nil {
		fmt.Fprintln(os.Stderr, red(BLANKS, "create error: ", err))
	}

	fmt.Println(green(BLANKS, "Queue ", q.Name, " has been successfully created."))
	if tag, err := getHudTag(q.Settings); err == nil {
		fmt.Printf("%sVisit hud-e.iron.io/mq/%s/projects/%s/queues/%s to see your queue.\n", BLANKS,
			tag,
			q.Settings.ProjectId,
			q.Name)
	}
}
func (r *RmCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq remove QUEUE_NAME

    Delete a queue from a project
    `)
	}
}

func (r *RmCmd) Flags(args ...string) error {
	r.flags = NewMqFlagSet(r.Usage())
	if err := r.flags.Parse(args); err != nil {
		return err
	}
	return nil
}

func (r *RmCmd) Args() error {
	if r.flags.NArg() < 1 && !isPipedIn() {
		return errors.New("rm requires a queue name.")
	}

	r.queue_name = r.flags.Arg(0)
	return nil
}
func (r *RmCmd) Run() {
	var queues []mq.Queue

	if isPipedIn() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			name := scanner.Text()
			queues = append(queues, mq.ConfigNew(name, &r.settings))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		queues = append(queues, mq.ConfigNew(r.queue_name, &r.settings))
	}

	for _, q := range queues {
		err := q.Delete()
		if err != nil {
			fmt.Println(red(BLANKS, "Error deleting queue", q.Name, ":", err))
		} else {
			fmt.Println(green(BLANKS, q.Name, " has been sucessfully deleted."))
		}
	}
}

func (c *ClearCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, "usage: iron mq clear QUEUE_NAME")
	}
}

func (c *ClearCmd) Flags(args ...string) error {
	c.flags = NewMqFlagSet(c.Usage())

	if err := c.flags.Parse(args); err != nil {
		return err
	}
	return nil
}

func (c *ClearCmd) Args() error {
	if c.flags.NArg() < 1 {
		return errors.New(`clear requires one arg

    usage: iron mq clear QUEUE_NAME`)
	}
	c.queue_name = c.flags.Arg(0)
	return nil
}

func (c *ClearCmd) Run() {
	q := mq.ConfigNew(c.queue_name, &c.settings)
	if err := q.Clear(); err != nil {
		fmt.Println(red(BLANKS, "Error clearing queue:", err))
		return
	}
	fmt.Fprintln(os.Stderr, green(BLANKS, "Queue", q.Name, "has been successfully cleared"))
}

func (p *PeekCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq peek [--n number] QUEUE_NAME

    n: peek n numbers of messages(default: 1, max: 100)`)
	}
}

func (p *PeekCmd) Flags(args ...string) error {
	p.flags = NewMqFlagSet(p.Usage())
	p.n = p.flags.n()

	if err := p.flags.Parse(args); err != nil {
		return err
	}

	return p.flags.validateAllFlags()
}

func (p *PeekCmd) Args() error {
	if p.flags.NArg() < 1 {
		return errors.New(`peek requires one arg

    usage: iron mq peek [--n numer] QUEUE_NAME`)
	}
	p.queue_name = p.flags.Arg(0)
	return nil
}

func (p *PeekCmd) Run() {
	q := mq.ConfigNew(p.queue_name, &p.settings)

	msgs, err := q.PeekN(*p.n)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(BLANKS, "Clear error:", err))
		return
	}

	if !isPipedOut() {
		fmt.Println("-------- ID ------ | Body")
	}
	printMessages(msgs)
}

func (p *PushCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq push [-f file] QUEUE_NAME "MESSAGE" "MESSAGE"...

    f: json file with message bodies to be used. Format should be '{"messages": ["1", "2", "3"...]}'`)
	}
}

func (p *PushCmd) Flags(args ...string) error {
	p.flags = NewMqFlagSet(p.Usage())

	p.filename = p.flags.filename()

	if err := p.flags.Parse(args); err != nil {
		return err
	}

	return p.flags.validateAllFlags()
}

func (p *PushCmd) Args() error {
	if p.flags.NArg() < 1 && !isPipedIn() {
		return errors.New(`push requires the queue name

    usage: iron mq push [-f file] QUEUE_NAME "MESSAGE"...`)
	}

	p.queue_name = p.flags.Arg(0)

	if *p.filename != "" {
		b, err := ioutil.ReadFile(*p.filename)
		if err != nil {
			return err
		}

		messageStruct := struct {
			Messages []string `json:"messages"`
		}{}
		err = json.Unmarshal(b, &messageStruct)
		if err != nil {
			return err
		}

		p.messages = append(p.messages, messageStruct.Messages...)
	}

	if p.flags.NArg() > 1 {
		p.messages = append(p.messages, p.flags.Args()[1:]...)
	}

	if len(p.messages) < 1 {
		return errors.New(`push requires at least one message

    usage: iron mq push [-f file] QUEUE_NAME "MESSAGE" "MESSAGE 2"...`)
	}
	return nil
}

func (p *PushCmd) Run() {
	q := mq.ConfigNew(p.queue_name, &p.settings)

	ids, err := q.PushStrings(p.messages...)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err))
	}

	if isPipedOut() {
		for _, id := range ids {
			fmt.Println(id)
		}
	} else {
		fmt.Println(green(LINES, "Message succesfully pushed!"))
		fmt.Printf("%sMessage IDs:\n", BLANKS)
		fmt.Printf("%s", BLANKS)
		for _, id := range ids {
			fmt.Printf("%s ", id)
		}
		fmt.Println()
		if tag, err := getHudTag(q.Settings); err == nil {
			fmt.Printf("%sGo to hud-e.iron.io/mq/%s/projects/%s/queues/%s for more info\n", BLANKS,
				tag, q.Settings.ProjectId, q.Name)
		}
	}
}

func (p *PopCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq pop [-n int] [-o file] QUEUE_NAME

    pop reserves then deletes a message from the queue
    n: number of messages to pop off the queue, default: 1
    o: write results in json to a file`)
	}
}

func (p *PopCmd) Flags(args ...string) error {
	p.flags = NewMqFlagSet(p.Usage())

	p.n = p.flags.n()
	p.outputfile = p.flags.outputfile()

	if err := p.flags.Parse(args); err != nil {
		return err
	}
	return p.flags.validateAllFlags()
}

func (p *PopCmd) Args() error {
	if p.flags.NArg() < 1 {
		return errors.New(`pop requires a queue name

    usage: iron mq pop [-n n] [-o file] QUEUE_NAME`)
	}
	if *p.outputfile != "" {
		f, err := os.Create(*p.outputfile)
		if err != nil {
			return err
		}
		p.file = f
	}

	p.queue_name = p.flags.Arg(0)
	return nil
}

func (p *PopCmd) Run() {
	q := mq.ConfigNew(p.queue_name, &p.settings)

	messages, err := q.PopN(*p.n)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err))
	}

	// If anything here fails, we still want to print out what was deleted before exiting
	if p.file != nil {
		b, err := json.Marshal(messages)
		if err != nil {
			fmt.Fprintln(os.Stderr, red(err))
			printMessages(messages)
		}
		_, err = p.file.Write(b)
		if err != nil {
			fmt.Fprintln(os.Stderr, red(err))
			printMessages(messages)
		}
	}

	if isPipedOut() {
		printMessages(messages)
	} else {
		fmt.Println(green("messages successfully popped off the queue"))
		fmt.Println("-------- ID ------ | Body")
		printMessages(messages)
	}
}

func (r *ReserveCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq reserve [-t timeout] [-n n] [-o file] QUEUE_NAME

    t: timeout until message is put back on the queue, default: 60
    n: number of messages to reserve
    o: write results in json to a file`)
	}
}

func (r *ReserveCmd) Flags(args ...string) error {
	r.flags = NewMqFlagSet(r.Usage())

	r.n = r.flags.n()
	r.timeout = r.flags.timeout()
	r.outputfile = r.flags.outputfile()

	if err := r.flags.Parse(args); err != nil {
		return err
	}
	return r.flags.validateAllFlags()

}
func (r *ReserveCmd) Args() error {
	if r.flags.NArg() < 1 {
		return errors.New(`reserve requires a queue name

    usage: iron mq reserve [-t timeout] [-n n] [-o file] QUEUE_NAME`)
	}
	if *r.outputfile != "" {
		f, err := os.Create(*r.outputfile)
		if err != nil {
			return err
		}
		r.file = f
	}

	r.queue_name = r.flags.Arg(0)
	return nil
}

func (r *ReserveCmd) Run() {
	q := mq.ConfigNew(r.queue_name, &r.settings)
	messages, err := q.GetNWithTimeout(*r.n, *r.timeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err))
	}

	// If anything here fails, we still want to print out what was deleted before exiting
	if r.file != nil {
		b, err := json.Marshal(messages)
		if err != nil {
			fmt.Fprintln(os.Stderr, red(err))
			printMessages(messages)
			return
		}
		_, err = r.file.Write(b)
		if err != nil {
			fmt.Fprintln(os.Stderr, red(err))
			printMessages(messages)
			return
		}
	}

	if isPipedOut() {
		printMessages(messages)
	} else {
		fmt.Println(green("messages successfully reserved"))
		fmt.Println("-------- ID ------ | Body")
		printMessages(messages)
	}

}

type DeleteCmd struct {
	mqCmd

	filequeue_name *string
	queue_name     string
	ids            []string
}

func (d *DeleteCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq delete [-i file] QUEUE_NAME "MSG_ID" "MSG_ID"...

    Delete a message of a queue
    -i: json file with a set of ids to be deleted. Format should be {"ids": ["123", "456", ...]}`)
	}
}

func (d *DeleteCmd) Flags(args ...string) error {
	d.flags = NewMqFlagSet(d.Usage())

	d.filequeue_name = d.flags.filename()

	if err := d.flags.Parse(args); err != nil {
		return err
	}
	return d.flags.validateAllFlags()
}

func (d *DeleteCmd) Args() error {
	if d.flags.NArg() < 1 {
		usage := d.Usage()
		usage()
		return errors.New(`delete requires a queue name`)
	}
	d.queue_name = d.flags.Arg(0)

	// Read and parse piped info
	if isPipedIn() {
		ids, err := readIds()
		if err != nil {
			return err
		}
		d.ids = append(d.ids, ids...)
	}

	if *d.filequeue_name != "" {
		b, err := ioutil.ReadFile(*d.filequeue_name)
		if err != nil {
			return err
		}

		// Use the message struct so its compatible with output files from reserve
		var msgs []mq.Message
		err = json.Unmarshal(b, &msgs)
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			d.ids = append(d.ids, msg.Id)
		}
	}

	if d.flags.NArg() > 1 {
		d.ids = append(d.ids, d.flags.Args()[1:]...)
	}

	if len(d.ids) < 1 {
		return errors.New("delete requires at least one message id")
	}
	return nil
}

func (d *DeleteCmd) Run() {
	q := mq.ConfigNew(d.queue_name, &d.settings)

	err := q.DeleteMessages(d.ids)
	if err != nil {
		fmt.Fprintln(os.Stderr, red("Error deleting message", err))
	}
	fmt.Println(green("done deleting messages"))
}

type InfoCmd struct {
	mqCmd

	queue_name string
}

func (i *InfoCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq info QUEUE_NAME`)
	}
}

func (i *InfoCmd) Flags(args ...string) error {
	i.flags = NewMqFlagSet(i.Usage())

	if err := i.flags.Parse(args); err != nil {
		return err
	}

	return i.flags.validateAllFlags()
}

func (i *InfoCmd) Args() error {
	if i.flags.NArg() < 1 {
		return errors.New(`info requires a queue name`)
	}

	i.queue_name = i.flags.Arg(0)
	return nil
}

func (i *InfoCmd) Run() {
	q := mq.ConfigNew(i.queue_name, &i.settings)
	info, err := q.Info()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if tag, err := getHudTag(q.Settings); err == nil {
		fmt.Printf("Go to hud-e.iron.io/mq/%s/projects/%s/queues/%s for more info.\n", tag, q.Settings.ProjectId, info.Name)
	}
	fmt.Println(BLANKS, "Name:", info.Name)
	fmt.Println(BLANKS, "Current Size:", info.Size)
	fmt.Println(BLANKS, "Total messages:", info.TotalMessages)
	fmt.Println(BLANKS, "Message expiration:", info.MessageExpiration)
	fmt.Println(BLANKS, "Message timeout:", info.MessageTimeout)
	if info.Push != nil {
		fmt.Println(BLANKS, "type:", *info.Type)
		fmt.Println(BLANKS, "subscribers:", len(info.Push.Subscribers))
		fmt.Println(BLANKS, "retries:", info.Push.Retries)
		fmt.Println(BLANKS, "retries delay:", info.Push.RetriesDelay)
	}
}

// TODO: Figure out the region for the hud url
// seriously though
// this is super duper hacky
// it only works with the public cluster mq-aws-us-east-1-1
func getHudTag(settings config.Settings) (string, error) {
	res, err := http.Get("https://auth.iron.io/1/clusters?oauth=" + settings.Token)
	if err != nil {
		return "", err
	}

	clusters := struct {
		Clusters []struct {
			Tag string `json:"tag"`
			URL string `json:"url"`
		} `json:"clusters"`
	}{}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	err = json.Unmarshal(b, &clusters)
	if err != nil {
		return "", err
	}
	queueHost := settings.Host
	for _, cluster := range clusters.Clusters {
		if cluster.URL == queueHost {
			return cluster.Tag, err
		}
	}
	return "", fmt.Errorf("no hud tags found")
}
